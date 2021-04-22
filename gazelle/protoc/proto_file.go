package protoc

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/emicklei/proto"
)

func isProtoFile(filename string) bool {
	return filepath.Ext(filename) == ".proto"
}

// NewProtoFile takes the package directory and base name of the file (e.g.
// 'foo.proto') and constructs ProtoFile
func NewProtoFile(dir, basename string) *ProtoFile {
	return &ProtoFile{
		Dir:      dir,
		Basename: basename,
		Name:     strings.TrimSuffix(basename, filepath.Ext(basename)),
	}
}

// ProtoFile represents a proto file we discover in a package.
type ProtoFile struct {
	Dir      string // e.g. "rosetta/rosetta/common/"
	Basename string // e.g. "foo.proto"
	Name     string // e.g. "foo"

	protoPackage *proto.Package
	imports      []*proto.Import
	services     []*proto.Service
	messages     []*proto.Message
	options      []*proto.Option
	enums        []*proto.Enum
	enumOptions  []*proto.Option
}

// Relname returns the relative path of the proto file.
func (f *ProtoFile) Relname() string {
	if f.Dir == "" {
		return f.Basename
	}
	return filepath.Join(f.Dir, f.Basename)
}

// GetOptions returns the list of top-level options defined in the proto file.
func (f *ProtoFile) GetOptions() []*proto.Option {
	return f.options
}

// HasEnums returns true if the proto file has at least one enum.
func (f *ProtoFile) HasEnums() bool {
	return len(f.enums) > 0
}

// HasMessages returns true if the proto file has at least one message.
func (f *ProtoFile) HasMessages() bool {
	return len(f.messages) > 0
}

// HasServices returns true if the proto file has at least one service.
func (f *ProtoFile) HasServices() bool {
	return len(f.services) > 0
}

// HasEnumOption returns true if the proto file has at least one enum or enum
// field annotated with the given named field extension.
func (f *ProtoFile) HasEnumOption(name string) bool {
	for _, option := range f.enumOptions {
		if option.Name == name {
			return true
		}
	}
	return false
}

// Parse the source and walk the statements in the file.
func (f *ProtoFile) Parse() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not parse: %v", err)
	}

	if bwd, ok := os.LookupEnv("BUILD_WORKSPACE_DIRECTORY"); ok {
		wd = bwd
	}

	filename := filepath.Join(wd, f.Dir, f.Basename)
	reader, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("could not open %s: %w (cwd=%s)", filename, err, wd)
	}
	defer reader.Close()

	return f.parseReader(reader)
}

func (f *ProtoFile) parseReader(in io.Reader) error {
	parser := proto.NewParser(in)
	definition, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("could not parse %s/%s: %w", f.Dir, f.Basename, err)
	}

	proto.Walk(definition,
		proto.WithPackage(f.handlePackage),
		proto.WithOption(f.handleOption),
		proto.WithImport(f.handleImport),
		proto.WithService(f.handleService),
		proto.WithMessage(f.handleMessage),
		proto.WithEnum(f.handleEnum))

	// NOTE: f.options only holds top-level options.  To introspect the enum and
	// enum field options we need to do extra work.
	collector := &protoEnumOptionCollector{}
	for _, enum := range f.enums {
		collector.VisitEnum(enum)
	}
	f.enumOptions = collector.options

	return nil
}

func (f *ProtoFile) handlePackage(p *proto.Package) {
	f.protoPackage = p
}

func (f *ProtoFile) handleOption(o *proto.Option) {
	f.options = append(f.options, o)
}

func (f *ProtoFile) handleImport(i *proto.Import) {
	f.imports = append(f.imports, i)
}

func (f *ProtoFile) handleEnum(i *proto.Enum) {
	f.enums = append(f.enums, i)
}

func (f *ProtoFile) handleService(s *proto.Service) {
	f.services = append(f.services, s)
}

func (f *ProtoFile) handleMessage(m *proto.Message) {
	f.messages = append(f.messages, m)
}

// getGoPackageOption is a utility function to seek for the go_package option
// and split it.  If present the return values will be populated with the
// importpath and alias (e.g. github.com/foo/bar/v1;bar ->
// "github.com/foo/bar/v1", "bar").  If the option was not found the bool return
// argument is false.
func getGoPackageOption(options []*proto.Option) (string, string, bool) {
	for _, opt := range options {
		if opt.Name != "go_package" {
			continue
		}
		parts := strings.SplitN(opt.Constant.Source, ";", 2)
		switch len(parts) {
		case 0:
			return "", "", true
		case 1:
			return parts[0], "", true
		case 2:
			return parts[0], parts[1], true
		default:
			return parts[0], strings.Join(parts[1:], ";"), true
		}
	}

	return "", "", false
}

func matchingFiles(files map[string]*ProtoFile, srcs []label.Label) []*ProtoFile {
	log.Printf("matching %v files in %v", srcs, files)

	matching := make([]*ProtoFile, 0)
	for _, src := range srcs {
		if file, ok := files[src.Name]; ok {
			matching = append(matching, file)
		}
	}

	log.Printf("matched %d", len(matching))

	return matching
}
