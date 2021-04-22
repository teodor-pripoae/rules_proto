package protoc

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/bazelbuild/buildtools/build"
)

// ProtoCompileRule implements a ruleProvider for the @build_stack_rules_proto
// family of rules.  These all share a similar pattern but differ on naming
// based on the target language, whether or not it's a proto (only messages) or
// gRPC (has_services), and whether it's the "_compile" rule (which generates
// the sources) or the language-wrapped "_library" rule.
type ProtoCompileRule struct {
	prefix           string
	library          ProtoLibrary
	plugins          []label.Label
	generatedSrcs    []string
	generatedOptions map[string][]string
	visibility       []string
	comment          []string
}

// NewProtoCompileRule constructs a new ProtoCompileRule based on the proto_library on which
// it depends, as well as the precomputed list of GeneratedSrcs.
func NewProtoCompileRule(
	prefix string,
	library ProtoLibrary,
	plugins []label.Label,
	generatedSrcs []string,
	generatedOptions map[string][]string) *ProtoCompileRule {
	rule := &ProtoCompileRule{
		prefix:           prefix,
		library:          library,
		plugins:          plugins,
		generatedSrcs:    generatedSrcs,
		generatedOptions: generatedOptions,
	}
	return rule
}

// Kind implements part of the ruleProvider interface.
func (s *ProtoCompileRule) Kind() string {
	return fmt.Sprintf("proto_compile")
}

// Name implements part of the ruleProvider interface.
func (s *ProtoCompileRule) Name() string {
	return fmt.Sprintf("%s_%s_compile", s.library.BaseName(), s.prefix)
}

// Imports implements part of the ruleProvider interface.
func (s *ProtoCompileRule) Imports() []string {
	return []string{s.Kind()}
}

// Visibility implements part of the ruleProvider interface.
func (s *ProtoCompileRule) Visibility() []string {
	return s.visibility
}

// Rule implements part of the ruleProvider interface.
func (s *ProtoCompileRule) Rule() *rule.Rule {
	newRule := rule.NewRule(s.Kind(), s.Name())
	visibility := s.Visibility()
	if len(s.visibility) > 0 {
		newRule.SetAttr("visibility", visibility)
	}
	if s.comment != nil {
		for _, line := range s.comment {
			newRule.AddComment(line)
		}
	}

	newRule.SetAttr("proto", s.library.Name())
	newRule.SetAttr("plugins", s.pluginLabels())
	newRule.SetAttr("generated_srcs", s.GeneratedSrcs())

	if len(s.generatedOptions) > 0 {
		newRule.SetAttr("options", s.Options())
	}

	// // special case for go_package option.  TODO: refactor this to make a
	// // subclass like go_proto_rule that does this.
	// if strings.HasPrefix(s.lang, "go") {
	// 	for _, file := range s.library.Files() {
	// 		pkg, _, ok := getGoPackageOption(file.GetOptions())
	// 		if ok {
	// 			newRule.SetAttr("go_package", pkg)
	// 			break
	// 		}
	// 	}
	// }
	return newRule
}

// KindInfo implements part of the ruleProvider interface.
func (s *ProtoCompileRule) KindInfo() rule.KindInfo {
	return rule.KindInfo{
		NonEmptyAttrs:  map[string]bool{"deps": true},
		MergeableAttrs: map[string]bool{},
	}
}

// Deps computes the dependencies of the rule.
func (s *ProtoCompileRule) Deps() []string {
	return []string{":" + s.library.Name()}
}

// pluginLabels returns the label strings for the plugins.
func (s *ProtoCompileRule) pluginLabels() []string {
	labels := make([]string, len(s.plugins))
	for i, lab := range s.plugins {
		labels[i] = lab.String()
	}
	return labels
}

// GeneratedSrcs computes the source files that are generated by the rule.  The
// implementation currently hardcodes the information that is encapsulated by
// the @build_stack_rules_proto "proto_plugin" provider.
func (s *ProtoCompileRule) GeneratedSrcs() []string {
	return s.generatedSrcs
}

// Options computes the options string_list_dict.
func (s *ProtoCompileRule) Options() build.Expr {
	items := make([]*build.KeyValueExpr, 0)
	keys := make([]string, 0)
	for k := range s.generatedOptions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		opts := s.generatedOptions[key]
		sort.Strings(opts)
		value := &build.StringExpr{Value: strings.Join(opts, ",")}
		items = append(items, &build.KeyValueExpr{
			Key:   &build.StringExpr{Value: key},
			Value: value,
		})
	}
	return &build.DictExpr{List: items}
}

// Resolve implements part of the RuleProvider interface.
func (s *ProtoCompileRule) Resolve(c *config.Config, r *rule.Rule, importsRaw interface{}, from label.Label) {
}