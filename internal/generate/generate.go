package generate

import (
	"bytes"
	"text/template"

	"github.com/autohttp/autohttp/internal/analyze"
	"github.com/autohttp/autohttp/session"
	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
)

type templateData struct {
	SessionID     string
	TargetURL     string
	ExchangeCount int
	Dependencies  []*analyze.Dependency
}

func GoScript(sess *session.Session, analysis *analyze.Result) ([]byte, error) {
	data := templateData{
		SessionID:     sess.ID,
		TargetURL:     sess.TargetURL,
		ExchangeCount: len(sess.Exchanges),
		Dependencies:  analysis.Dependencies,
	}

	tmpl, err := template.New("script").Parse(GoTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(graph *pb.ExecutionGraph, target string) ([]byte, error) {
	return nil, nil
}
