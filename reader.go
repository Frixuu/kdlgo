package kdl

import (
	"bufio"
	"bytes"
)

type reader struct {
	reader *bufio.Reader
	line   int
	pos    int
	depth  int
}

func wrapReader(r *bufio.Reader) reader {
	return reader{reader: r, line: 1, pos: 0}
}

func (r *reader) readRune() (ch rune, err error) {

	ch, _, err = r.reader.ReadRune()
	if err != nil {
		return
	}

	if isNewLine(ch) {

		if ch == '\r' {

			next, _, errNext := r.reader.ReadRune()
			if errNext != nil {
				return
			}

			if next == '\n' {
				_ = r.reader.UnreadRune()
				return
			}
		}

		r.line++
		r.pos = 0
		return
	}

	r.pos++
	return
}

func (r *reader) discardRunes(count int) {
	for i := 0; i < count; i++ {
		_, _ = r.readRune()
	}
}

func (r *reader) readByte() (b byte, err error) {
	b, err = r.reader.ReadByte()
	if b == '\n' || b == '\r' {
		r.line++
		r.pos = 0
	} else {
		r.pos++
	}
	return
}

func (r *reader) peekByte() (b byte, err error) {
	b, err = r.reader.ReadByte()
	if err != nil {
		return
	}
	err = r.reader.UnreadByte()
	return
}

func (r *reader) discardByte() {
	_, _ = r.readByte()
}

func (r *reader) readBytes(count int) (bytes []byte, err error) {

	bytes = make([]byte, count)
	wasCR := false

	for count > 0 {

		var n int
		n, err = r.reader.Read(bytes)
		if err != nil {
			return
		}

		for i := 0; i < n; i++ {

			b := bytes[i]
			if b == '\n' && !wasCR {
				r.line++
				r.pos = 0
				wasCR = false
				continue
			}

			if b == '\r' {
				wasCR = true
				r.line++
				r.pos = 0
				continue
			}

			wasCR = false
			r.pos++
		}

		count -= n
	}

	return
}

func (r *reader) discardBytes(count int) {

	bytes, err := r.peekBytes(count)
	if err != nil {
		_, _ = r.readBytes(count)
		return
	}

	wasCR := false
	for _, b := range bytes {

		if b == '\n' && !wasCR {
			r.line++
			r.pos = 0
			wasCR = false
			continue
		}

		if b == '\r' {
			wasCR = true
			r.line++
			r.pos = 0
			continue
		}

		wasCR = false
		r.pos++
	}

	r.reader.Discard(count)
}

// peekBytes tries to return next N bytes without advancing the reader.
func (r *reader) peekBytes(count int) ([]byte, error) {
	return r.reader.Peek(count)
}

func (r *reader) peekRune() (rune, error) {
	ch, _, err := r.reader.ReadRune()
	if err != nil {
		return ch, err
	}

	err = r.reader.UnreadRune()
	return ch, err
}

func (r *reader) isNext(expected []byte) (bool, error) {

	next, err := r.peekBytes(len(expected))
	if err != nil {
		return false, err
	}

	return bytes.Equal(next, expected), nil
}
