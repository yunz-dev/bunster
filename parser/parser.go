package parser

import (
	"fmt"

	"github.com/yassinebenaid/bunny/ast"
	"github.com/yassinebenaid/bunny/lexer"
	"github.com/yassinebenaid/bunny/token"
)

func New(l lexer.Lexer) Parser {
	var p = Parser{l: l}

	// So that both curr and next tokens get initialized.
	p.proceed()
	p.proceed()

	return p
}

type Parser struct {
	l     lexer.Lexer
	curr  token.Token
	next  token.Token
	Error error
}

func (p *Parser) error(msg string, args ...any) {
	if p.Error == nil {
		p.Error = fmt.Errorf("syntax error: "+msg+".", args...)
	}
}

func (p *Parser) proceed() {
	p.curr = p.next
	p.next = p.l.NextToken()
}

func (p *Parser) ParseScript() ast.Script {
	var script ast.Script

	for ; p.curr.Type != token.EOF; p.proceed() {
		switch p.curr.Type {
		case token.BLANK, token.NEWLINE:
			continue
		case token.HASH:
			for p.curr.Type != token.NEWLINE && p.curr.Type != token.EOF {
				p.proceed()
			}
		default:
			script.Statements = append(script.Statements, p.parseCommandList())
		}
	}

	return script
}

func (p *Parser) parseCommandList() ast.Node {
	var left ast.Node
	pipe := p.parsePipline()
	if len(pipe) == 1 {
		left = pipe[0].Command
	} else {
		left = pipe
	}

	for p.curr.Type == token.AND || p.curr.Type == token.OR {
		operator := p.curr.Literal
		p.proceed()
		for p.curr.Type == token.BLANK || p.curr.Type == token.NEWLINE {
			p.proceed()
		}

		var right ast.Node
		rightPipe := p.parsePipline()
		if len(rightPipe) == 1 {
			right = rightPipe[0].Command
		} else {
			right = rightPipe
		}

		left = ast.BinaryConstruction{
			Left:     left,
			Operator: operator,
			Right:    right,
		}
	}

	if p.curr.Type == token.AMPERSAND {
		return ast.BackgroundConstruction{Node: left}
	}

	return left
}

func (p *Parser) parsePipline() ast.Pipeline {
	var pipeline ast.Pipeline

	cmd := p.parseCommand()
	if cmd == nil {
		p.error("invalid command construction")
	}
	pipeline = append(pipeline, ast.PipelineCommand{Command: cmd})

	for {
		if p.curr.Type != token.PIPE && p.curr.Type != token.PIPE_AMPERSAND {
			break
		}
		var pipe ast.PipelineCommand
		pipeMethod := p.curr.Literal
		pipe.Stderr = p.curr.Type == token.PIPE_AMPERSAND

		p.proceed()
		for p.curr.Type == token.BLANK || p.curr.Type == token.NEWLINE {
			p.proceed()
		}

		pipe.Command = p.parseCommand()
		if pipe.Command == nil {
			p.error("invalid pipeline construction, a command is missing after `%s`", pipeMethod)
		}
		pipeline = append(pipeline, pipe)
	}

	return pipeline
}

func (p *Parser) parseCommand() ast.Node {
	if compound := p.getCompoundParser(); compound != nil {
		return compound()
	}

	var cmd ast.Command
	cmd.Name = p.parseField()
	if cmd.Name == nil {
		p.error("`%s` is a special character and cannot be used as a command name", p.curr.Literal)
		return nil
	}

loop:
	for {
		switch {
		case p.curr.Type == token.BLANK:
			break
		case p.curr.Type == token.EOF:
			break loop
		case p.curr.Type == token.HASH:
			for p.curr.Type != token.NEWLINE && p.curr.Type != token.EOF {
				p.proceed()
			}
		case p.isRedirectionToken():
			p.HandleRedirection(&cmd.Redirections)
		default:
			if p.isControlToken() {
				break loop
			}

			cmd.Args = append(cmd.Args, p.parseField())
		}

		if !p.isRedirectionToken() && !p.isControlToken() {
			p.proceed()
		}
	}
	return cmd
}

func (p *Parser) parseField() ast.Node {
	var nodes []ast.Node

loop:
	for {
		switch p.curr.Type {
		case token.BLANK, token.EOF:
			break loop
		case token.SIMPLE_EXPANSION:
			nodes = append(nodes, ast.SimpleExpansion(p.curr.Literal))
		case token.SINGLE_QUOTE:
			nodes = append(nodes, p.parseLiteralString())
		case token.DOUBLE_QUOTE:
			nodes = append(nodes, p.parseString())
		case token.INT:
			if len(nodes) == 0 && p.isRedirectionToken() {
				break loop
			}
			fallthrough
		default:
			if p.curr.Type != token.INT && p.isRedirectionToken() || p.isControlToken() {
				break loop
			}

			nodes = append(nodes, ast.Word(p.curr.Literal))
		}

		p.proceed()
	}

	return concat(nodes)
}

func (p *Parser) parseLiteralString() ast.Word {
	p.proceed()

	if p.curr.Type == token.SINGLE_QUOTE {
		return ast.Word("")
	}

	word := p.curr.Literal
	p.proceed()

	if p.curr.Type != token.SINGLE_QUOTE {
		p.error("a closing single quote is missing")
	}

	return ast.Word(word)
}

func (p *Parser) parseString() ast.Node {
	p.proceed()

	if p.curr.Type == token.DOUBLE_QUOTE {
		return ast.Word("")
	}

	var nodes []ast.Node

loop:
	for {
		switch p.curr.Type {
		case token.DOUBLE_QUOTE, token.EOF:
			break loop
		case token.ESCAPED_CHAR:
			nodes = append(nodes, ast.Word("\\"+p.curr.Literal))
		case token.SIMPLE_EXPANSION:
			nodes = append(nodes, ast.SimpleExpansion(p.curr.Literal))
		default:
			nodes = append(nodes, ast.Word(p.curr.Literal))
		}

		p.proceed()
	}

	if p.curr.Type != token.DOUBLE_QUOTE {
		p.error("a closing double quote is missing")
	}

	return concat(nodes)
}

func concat(n []ast.Node) ast.Node {
	var conc ast.Concatination
	var mergedWords ast.Word
	var hasWords bool

	for i, node := range n {

		if w, ok := node.(ast.Word); ok {
			mergedWords += w
			hasWords = true
		} else {
			if hasWords {
				conc.Nodes = append(conc.Nodes, mergedWords)
				mergedWords, hasWords = "", false
			}
			conc.Nodes = append(conc.Nodes, node)
		}

		if i == len(n)-1 && hasWords {
			conc.Nodes = append(conc.Nodes, mergedWords)
		}
	}

	if len(conc.Nodes) == 0 {
		return nil
	}

	if len(conc.Nodes) == 1 {
		return conc.Nodes[0]
	}

	return conc
}

func (p *Parser) isControlToken() bool {
	return p.curr.Type == token.PIPE ||
		p.curr.Type == token.PIPE_AMPERSAND ||
		p.curr.Type == token.AND ||
		p.curr.Type == token.OR ||
		p.curr.Type == token.AMPERSAND ||
		p.curr.Type == token.SEMICOLON ||
		p.curr.Type == token.NEWLINE
}
