package shell

import (
	"io"

	"github.com/codecrafters-io/shell-starter-go/assert"
)

type stateFunc func(*parser) stateFunc

type parser struct {
	state  stateFunc
	reader io.ByteScanner
	err    error
}

func parseInpust(reader io.ByteScanner) ([]string, error) {
	assert.NotNil(reader)

	p := &parser{
		state:  parseInputStart,
		reader: reader,
	}
	for p.state != nil {
		p.state = p.state(p)
	}
	return nil, nil
}

func (p *parser) accept(valid string) error {
	assert.NotNil(p.reader)

	for {
		_, err := p.reader.ReadByte()
		if err != nil {
			p.err = err
			return err
		}

	}
}

func parseInputStart(p *parser) stateFunc {

	return nil
}
