package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	pb "github.com/autohttp/autohttp/gen/autohttp/v1"
	"github.com/autohttp/autohttp/internal/analyze"
	"github.com/autohttp/autohttp/internal/generate"
	"github.com/autohttp/autohttp/internal/importer"
	"github.com/autohttp/autohttp/internal/tree"
	"github.com/autohttp/autohttp/session"
)

const version = "0.1.0"

func main() {
	log.SetFlags(0)
	log.SetPrefix("autohttp: ")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "record":
		cmdRecord()
	case "stop":
		cmdStop()
	case "analyze":
		cmdAnalyze()
	case "inspect":
		cmdInspect()
	case "generate":
		cmdGenerate()
	case "verify":
		cmdVerify()
	case "version":
		fmt.Println("autohttp version", version)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "autohttp: unknown command %q\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: autohttp <command> [flags]

Commands:
  record <url>   Start a new recording session
  stop           Finalize the active recording
  analyze        Run deterministic analysis
  inspect        Inspect session artifacts
  generate       Generate a standalone script
  verify         Verify a generated script
  version        Print version
  help           Print this help`)
}

func cmdRecord() {
	fs := flag.NewFlagSet("record", flag.ExitOnError)
	noAI := fs.Bool("no-ai", false, "Disable AI escalation")
	fs.Parse(os.Args[2:])
	_ = noAI
	fmt.Println("record: not yet implemented")
}

func cmdStop() {
	fmt.Println("stop: not yet implemented")
}

func cmdAnalyze() {
	fs := flag.NewFlagSet("analyze", flag.ExitOnError)
	fs.StringVar(&sessionPath, "session", "", "Path to session fixture")
	noAI := fs.Bool("no-ai", false, "Disable AI escalation")
	fs.Parse(os.Args[2:])
	_ = noAI

	if sessionPath == "" {
		fmt.Fprintln(os.Stderr, "autohttp: --session is required")
		os.Exit(1)
	}

	sess, err := importer.ImportFixture(sessionPath)
	if err != nil {
		log.Fatalf("loading session: %v", err)
	}

	parser := tree.New()
	var trees []*tree.Tree
	for _, e := range sess.Exchanges {
		trees = append(trees, parser.Parse(exchangeToPB(e))...)
	}

	result := analyze.Analyze(sess, trees)

	fmt.Printf("Session: %s\n", sess.ID)
	fmt.Printf("Exchanges: %d\n", len(sess.Exchanges))
	fmt.Printf("Dependencies: %d\n", len(result.Dependencies))
	for _, dep := range result.Dependencies {
		fmt.Printf("  %s → %s (%s) [%.0f%%]\n", dep.From, dep.To, dep.Value, dep.Confidence*100)
	}
}

func cmdInspect() {
	fs := flag.NewFlagSet("inspect", flag.ExitOnError)
	fs.StringVar(&sessionPath, "session", "", "Path to session fixture")
	fs.Parse(os.Args[2:])

	if sessionPath == "" {
		fmt.Fprintln(os.Stderr, "autohttp: --session is required")
		os.Exit(1)
	}

	sess, err := importer.ImportFixture(sessionPath)
	if err != nil {
		log.Fatalf("loading session: %v", err)
	}

	fmt.Printf("Session ID: %s\n", sess.ID)
	fmt.Printf("Target URL: %s\n", sess.TargetURL)
	fmt.Printf("Exchange count: %d\n", len(sess.Exchanges))
	for _, e := range sess.Exchanges {
		method := ""
		url := ""
		status := 0
		if e.Request != nil {
			method = e.Request.Method
			url = e.Request.URL
		}
		if e.Response != nil {
			status = e.Response.Status
		}
		fmt.Printf("  %s %s %s → %d\n", e.ID, method, url, status)
	}
}

func cmdGenerate() {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	fs.StringVar(&sessionPath, "session", "", "Path to session fixture")
	fs.StringVar(&outputFile, "output", "", "Output file path")
	target := fs.String("target", "go", "Target language (go|python)")
	fs.Parse(os.Args[2:])
	_ = target

	if sessionPath == "" {
		fmt.Fprintln(os.Stderr, "autohttp: --session is required")
		os.Exit(1)
	}

	sess, err := importer.ImportFixture(sessionPath)
	if err != nil {
		log.Fatalf("loading session: %v", err)
	}

	parser := tree.New()
	var trees []*tree.Tree
	for _, e := range sess.Exchanges {
		trees = append(trees, parser.Parse(exchangeToPB(e))...)
	}

	result := analyze.Analyze(sess, trees)

	code, err := generate.GoScript(sess, result)
	if err != nil {
		log.Fatalf("generating script: %v", err)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, code, 0644); err != nil {
			log.Fatalf("writing output: %v", err)
		}
		fmt.Printf("Generated script written to %s\n", outputFile)
	} else {
		os.Stdout.Write(code)
	}
}

func cmdVerify() {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	script := fs.String("script", "", "Path to generated script")
	successURL := fs.String("success-url", "", "Expected success URL")
	fs.Parse(os.Args[2:])
	_ = script
	_ = successURL
	fmt.Println("verify: not yet implemented")
}

func exchangeToPB(e *session.Exchange) *pb.HttpExchange {
	pbe := &pb.HttpExchange{
		Id:          e.ID,
		StartedAt:   e.StartedAt.UnixMilli(),
		CompletedAt: e.CompletedAt.UnixMilli(),
		Initiator:   e.Initiator,
	}
	if e.Request != nil {
		pbe.Request = &pb.Request{
			Method:   e.Request.Method,
			Url:      e.Request.URL,
			Body:     e.Request.Body,
			BodyType: e.Request.BodyType,
			Headers:  make([]*pb.Header, 0, len(e.Request.Headers)),
			Cookies:  make([]*pb.CookieMutation, 0, len(e.Request.Cookies)),
		}
		for k, v := range e.Request.Headers {
			pbe.Request.Headers = append(pbe.Request.Headers, &pb.Header{Key: k, Value: v})
		}
		for k, v := range e.Request.Cookies {
			pbe.Request.Cookies = append(pbe.Request.Cookies, &pb.CookieMutation{Name: k, Value: v})
		}
	}
	if e.Response != nil {
		pbe.Response = &pb.Response{
			Status:     int32(e.Response.Status),
			StatusText: e.Response.StatusText,
			Body:       e.Response.Body,
			BodyType:   e.Response.BodyType,
			Headers:    make([]*pb.Header, 0, len(e.Response.Headers)),
			SetCookies: make([]*pb.CookieMutation, 0, len(e.Response.SetCookies)),
		}
		for k, v := range e.Response.Headers {
			pbe.Response.Headers = append(pbe.Response.Headers, &pb.Header{Key: k, Value: v})
		}
		for k, v := range e.Response.SetCookies {
			pbe.Response.SetCookies = append(pbe.Response.SetCookies, &pb.CookieMutation{Name: k, Value: v})
		}
	}
	return pbe
}
