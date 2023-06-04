package kdl

import (
	"bufio"
	"bytes"
)

type reader struct {
	reader  *bufio.Reader
	line    int
	pos     int
	current rune
}

func wrapReader(r *bufio.Reader) *reader {
	return &reader{reader: r, line: 1, pos: 0}
}

func (r *reader) readRune() (rune, error) {
	ch, _, err := r.reader.ReadRune()
	if ch == '\n' {
		r.line++
		r.pos = 0
	} else {
		r.pos++
	}

	if err == nil {
		r.current = ch
	}

	return ch, err
}

func (r *reader) lastRead() rune {
	return r.current
}

func (r *reader) discardLine() error {
	_, err := r.reader.ReadString('\n')
	if err != nil {
		return err
	}

	err = r.reader.UnreadByte()
	return err
}

func (r *reader) discard(count int) {
	s, _ := r.peekN(count)
	for _, b := range s {
		var nl byte = '\n'
		if b == nl {
			r.line++
			r.pos = 0
		} else {
			r.pos++
		}
	}
	r.reader.Discard(count)
}

func (r *reader) peekN(count int) ([]byte, error) {
	return r.reader.Peek(count)
}

func (r *reader) peek() (rune, error) {
	ch, _, err := r.reader.ReadRune()
	if err != nil {
		return ch, err
	}

	err = r.reader.UnreadRune()
	return ch, err
}

func (r *reader) unreadRune() error {
	err := r.reader.UnreadRune()
	if err != nil {
		return err
	}

	peek, _ := r.reader.Peek(1)
	var b byte = '\n'
	if peek[0] == b {
		r.line--
	} else {
		r.pos--
	}

	return nil
}

func (r *reader) isNext(expected []byte) (bool, error) {

	next, err := r.peekN(len(expected))
	if err != nil {
		return false, err
	}

	if bytes.Equal(next, expected) {
		r.discard(len(expected))
		return true, nil
	}

	return false, nil
}
