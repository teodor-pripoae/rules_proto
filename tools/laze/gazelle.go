// Package laze generates provider tests target for every `proto_plugin` and
// `proto_rule` rule in each package.
package laze

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

const (
	languageName = "rules_proto"
	debug        = false
)

var logger = initLog()

func initLog() *log.Logger {
	if debug {
		return nil
	}

	f, err := os.OpenFile("/tmp/gazelle-laze.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	// defer f.Close()

	return log.New(f, "gazelle-laze", log.LstdFlags)
}

type protoRuleLang struct{}

// NewLanguage is called by Gazelle to install this language extension in a binary.
func NewLanguage() language.Language {
	return &protoRuleLang{}
}

// Name returns the name of the language. This should be a prefix of the kinds
// of rules generated by the language, e.g., "go" for the Go extension since it
// generates "go_library" rules.
func (*protoRuleLang) Name() string { return languageName }

// The following methods are implemented to satisfy the
// https://pkg.go.dev/github.com/bazelbuild/bazel-gazelle/resolve?tab=doc#Resolver
// interface, but are otherwise unused.
func (*protoRuleLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {}
func (*protoRuleLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error          { return nil }
func (*protoRuleLang) KnownDirectives() []string                                    { return nil }
func (*protoRuleLang) Configure(c *config.Config, rel string, f *rule.File)         {}

// Kinds returns a map of maps rule names (kinds) and information on how to
// match and merge attributes that may be found in rules of those kinds. All
// kinds of rules generated for this language may be found here.
func (*protoRuleLang) Kinds() map[string]rule.KindInfo {
	return kinds
}

// Loads returns .bzl files and symbols they define. Every rule generated by
// GenerateRules, now or in the past, should be loadable from one of these
// files.
func (*protoRuleLang) Loads() []rule.LoadInfo {
	return []rule.LoadInfo{
		{
			Name:    "@build_stack_rules_proto//:proto_plugin_info_provider_test.bzl",
			Symbols: []string{"proto_plugin_info_provider_test"},
		},
		{
			Name:    "@build_stack_rules_proto//:proto_rule_info_provider_test.bzl",
			Symbols: []string{"proto_rule_info_provider_test"},
		},
		{
			Name:    "@build_stack_rules_proto//:proto_rule_test.bzl",
			Symbols: []string{"proto_rule_test"},
		},
	}
}

// Fix repairs deprecated usage of language-specific rules in f. This is called
// before the file is indexed. Unless c.ShouldFix is true, fixes that delete or
// rename rules should not be performed.
func (*protoRuleLang) Fix(c *config.Config, f *rule.File) {}

// Imports returns a list of ImportSpecs that can be used to import the rule r.
// This is used to populate RuleIndex.
//
// If nil is returned, the rule will not be indexed. If any non-nil slice is
// returned, including an empty slice, the rule will be indexed.
func (b *protoRuleLang) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	srcs := r.AttrStrings("srcs")
	imports := make([]resolve.ImportSpec, len(srcs))

	for i, src := range srcs {
		imports[i] = resolve.ImportSpec{
			// Lang is the language in which the import string appears (this
			// should match Resolver.Name).
			Lang: languageName,
			// Imp is an import string for the library.
			Imp: fmt.Sprintf("//%s:%s", f.Pkg, src),
		}
	}

	return imports
}

// Embeds returns a list of labels of rules that the given rule embeds. If a
// rule is embedded by another importable rule of the same language, only the
// embedding rule will be indexed. The embedding rule will inherit the imports
// of the embedded rule. Since SkyLark doesn't support embedding this should
// always return nil.
func (*protoRuleLang) Embeds(r *rule.Rule, from label.Label) []label.Label { return nil }

// Resolve translates imported libraries for a given rule into Bazel
// dependencies. Information about imported libraries is returned for each rule
// generated by language.GenerateRules in language.GenerateResult.Imports.
// Resolve generates a "deps" attribute (or the appropriate language-specific
// equivalent) for each import according to language-specific rules and
// heuristics.
func (*protoRuleLang) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, importsRaw interface{}, from label.Label) {
}

var kinds = map[string]rule.KindInfo{
	"proto_plugin_info_provider_test": {
		NonEmptyAttrs:  map[string]bool{"srcs": true, "deps": true},
		MergeableAttrs: map[string]bool{"srcs": true},
	},
	"proto_rule_info_provider_test": {
		NonEmptyAttrs:  map[string]bool{"srcs": true, "deps": true},
		MergeableAttrs: map[string]bool{"srcs": true},
	},
	"proto_rule_test": {
		NonEmptyAttrs:  map[string]bool{"srcs": true, "deps": true},
		MergeableAttrs: map[string]bool{"srcs": true},
	},
}

// GenerateRules extracts build metadata from source files in a directory.
// GenerateRules is called in each directory where an update is requested in
// depth-first post-order.
//
// args contains the arguments for GenerateRules. This is passed as a struct to
// avoid breaking implementations in the future when new fields are added.
//
// A GenerateResult struct is returned. Optional fields may be added to this
// type in the future.
//
// Any non-fatal errors this function encounters should be logged using
// log.Print.
func (*protoRuleLang) GenerateRules(args language.GenerateArgs) language.GenerateResult {
	var rules []*rule.Rule
	var imports []interface{}

	// Search the existing build file for proto_plugin rules
	for _, existingRule := range args.File.Rules {
		switch existingRule.Kind() {
		case "proto_plugin":
			providerTestName := existingRule.Name() + "_info_provider_test"
			r := rule.NewRule("proto_plugin_info_provider_test", providerTestName)
			r.SetAttr("srcs", []string{providerTestName + ".golden.1.prototext"})
			r.SetAttr("deps", []string{":" + existingRule.Name()})
			rules = append(rules, r)
			imports = append(imports, []string{"proto_plugin_info_provider_test"})
		case "proto_rule":
			providerTestName := existingRule.Name() + "_info_provider_test"
			r := rule.NewRule("proto_rule_info_provider_test", providerTestName)
			r.SetAttr("srcs", []string{providerTestName + ".golden.1.prototext"})
			r.SetAttr("deps", []string{":" + existingRule.Name()})
			rules = append(rules, r)
			imports = append(imports, []string{"proto_rule_info_provider_test"})

			ruleTestName := existingRule.Name() + "_test"
			r = rule.NewRule("proto_rule_test", ruleTestName)
			r.SetAttr("srcs", []string{existingRule.Name() + ".bzl", existingRule.Name() + "_deps.bzl"})
			r.SetAttr("deps", []string{":" + existingRule.Name()})
			rules = append(rules, r)
			imports = append(imports, []string{"proto_rule_test"})
		}

	}

	return language.GenerateResult{
		Gen:     rules,
		Imports: imports,
	}
}
