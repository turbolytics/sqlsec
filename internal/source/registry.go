package source

import (
	"net/http"

	"github.com/turbolytics/sqlsec/internal/source/github"
)

type Source string

const (
	GithubSource Source = "github"
	// Future: Add more sources like GitlabSource, BitbucketSource, etc.
)

type Validator interface {
	Validate(r *http.Request, secret string) error
}

type Parser interface {
	Parse(r *http.Request) (map[string]any, error)
	Type(r *http.Request) (string, error)
}

type Registry struct {
	sources    map[Source]any
	validators map[Source]Validator
	parsers    map[Source]Parser
}

func New() *Registry {
	return &Registry{
		sources:    make(map[Source]any),
		validators: make(map[Source]Validator),
		parsers:    make(map[Source]Parser),
	}
}

func (r *Registry) Init() {
	r.sources[GithubSource] = struct{}{}
	r.validators[GithubSource] = &github.GithubValidator{}
	r.parsers[GithubSource] = &github.GithubParser{}
}

func (r *Registry) IsEnabled(source Source) bool {
	_, ok := r.sources[source]
	return ok
}

func (r *Registry) GetValidator(source Source) Validator {
	return r.validators[source]
}

func (r *Registry) GetParser(source Source) Parser {
	return r.parsers[source]
}

var DefaultRegistry *Registry

func init() {
	DefaultRegistry = New()
	DefaultRegistry.Init()
}
