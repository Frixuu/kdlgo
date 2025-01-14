package kdl

import (
	"fmt"
	"io"
	"unicode/utf8"
)

var (
	errUnexpectedSemicolon    = fmt.Errorf("%w: unexpected ';' not terminating a node", ErrInvalidSyntax)
	errUnexpectedRightBracket = fmt.Errorf("%w: unexpected top-level '}'", ErrInvalidSyntax)
	errUnexpectedLineCont     = fmt.Errorf("%w: unexpected top-level '\\'", ErrInvalidSyntax)
	errUnexpectedSlashdash    = fmt.Errorf("%w: unexpected slashdash", ErrInvalidSyntax)
)

func readNodes(r *reader) (nodes []Node, err error) {

	nodes = make([]Node, 0, 3)

	for {
		for {
			err = readUntilSignificant(r, false)
			if err != nil {
				if err == io.EOF && r.depth == 0 {
					err = nil
				}
				return
			}

			var ch rune
			ch, err = r.peekRune()
			if err != nil {
				if err == io.EOF && r.depth == 0 {
					err = nil
				}
				return
			}

			if !isNewLine(ch) {
				if ch == ';' {
					err = errUnexpectedSemicolon
					return
				} else if ch == '}' {
					if r.depth == 0 {
						err = errUnexpectedRightBracket
					}
					r.discardByte()
					return
				} else if ch == '\\' {
					err = errUnexpectedLineCont
					return
				}
				break
			}

			err = skipUntilNewLine(r, true)
			if err != nil {
				return
			}
		}

		// A "slashdash" comment silences the whole node
		var slashdash bool
		slashdash, err = r.isNext(charsSlashDash[:])
		if err != nil {
			return
		}
		if slashdash {
			r.discardBytes(2)
		}

		err = readUntilSignificant(r, true)
		if err != nil {
			if err == io.EOF {
				err = errUnexpectedSlashdash
			}
			return
		}

		var node Node
		node, err = readNode(r)
		if err != nil {
			return
		}

		if !slashdash {
			nodes = append(nodes, node)
		}
	}
}

func readNode(r *reader) (Node, error) {

	node := NewNode("")

	hint, err := readMaybeTypeHint(r)
	if err != nil {
		return node, err
	}
	node.TypeHint = hint

	name, err, _ := readIdentifier(r, stopModeSemicolon)
	if err != nil {
		return node, err
	}

	node.Name = name

	for {

		err := readUntilSignificant(r, true)
		if err != nil {
			if err == io.EOF {
				return node, nil
			}
			return node, err
		}

		slashdash, err := r.isNext(charsSlashDash[:])
		if slashdash && err == nil {
			r.discardBytes(2)
		}

		err = readUntilSignificant(r, true)
		if err != nil {
			if err == io.EOF {
				return node, errUnexpectedSlashdash
			}
			return node, err
		}

		ch, err := r.peekRune()
		if err != nil {
			return node, err
		}

		if isNewLine(ch) {
			r.discardRunes(1)
			if slashdash {
				return node, errUnexpectedSlashdash
			}
			return node, nil
		} else if ch == ';' {
			r.discardByte()
			if slashdash {
				return node, errUnexpectedSlashdash
			}
			return node, nil
		} else if ch == '}' {
			if slashdash {
				return node, errUnexpectedSlashdash
			}
			return node, nil
		} else if ch == '{' {
			r.discardByte()
			r.depth++
			children, err := readNodes(r)
			if err != nil {
				return node, err
			}
			r.depth--
			if !slashdash {
				for i := range children {
					node.AddChild(children[i])
				}
			}
			r.discardByte()
		} else {
			err = readArgOrProp(r, &node, slashdash)
			if err != nil {
				return node, err
			}
		}
	}
}

var (
	errUnexpectedBareIdentifier       = fmt.Errorf("%w: unexpected bare identifier", ErrInvalidSyntax)
	errUnexpectedTokenAfterValue      = fmt.Errorf("%w: unexpected token after value", ErrInvalidSyntax)
	errUnexpectedTokenAfterIdentifier = fmt.Errorf("%w: unexpected token after identifier", ErrInvalidSyntax)
)

// readArgOrProp reads an argument or a property
// and adds them to the provided Node definition.
func readArgOrProp(r *reader, dest *Node, discard bool) error {

	hint, err := readMaybeTypeHint(r)
	if err != nil {
		return err
	}

	// This can only be a property if there is no type hint at this time
	if hint.IsAbsent() {
		i, err, quoted := readIdentifier(r, stopModeEquals)
		if err == nil {
			// Identifier read successfully.
			ch, err := r.peekRune()
			if err == io.EOF {
				if quoted {
					if !discard {
						dest.AddArg(NewStringValue(string(i), NoHint()))
					}
					return nil
				}
				return errUnexpectedBareIdentifier
			} else if err == nil {
				if isValidValueTerminator(ch) {
					if quoted {
						if !discard {
							dest.AddArgValue(NewStringValue(string(i), NoHint()))
						}
						return nil
					}
					return errUnexpectedBareIdentifier
				} else if ch == '=' {
					r.discardByte()
					v, err := readValue(r)
					if err != nil {
						return err
					}
					if !discard {
						dest.SetPropValue(i, v)
					}
					return nil
				}
				return errUnexpectedTokenAfterIdentifier
			}
			return err
		}

		// Else: Bad identifier. This should be a Value instead. Fallthrough.
	}

	v, err := readValue(r)
	if err != nil {
		// Not a valid Value
		return err
	}
	v.TypeHint = hint

	ch, err := r.peekRune()

	if err == io.EOF || (err == nil && isValidValueTerminator(ch)) {
		if !discard {
			dest.AddArg(v)
		}
		return nil
	} else if err != nil {
		return err
	}

	return errUnexpectedTokenAfterValue
}

// skipUntilNewLine discards the reader to the next new line character OR EOF.
//
// If afterBreak is true, the reader is positioned after the newline break.
// If it is false, the reader is positioned just before a newline rune. (singular, in case of CRLF)
func skipUntilNewLine(r *reader, afterBreak bool) error {

	for {

		// CRLF is a special case as it spans two runes, so we check it first
		if isCrlf, err := r.isNext(charsCRLF[:]); isCrlf && err == nil {
			if afterBreak {
				r.discardBytes(2)
			} else {
				// Leave the LF only to simplify later checks
				r.discardByte()
			}
			break
		}

		ch, err := r.peekRune()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if isNewLine(ch) {
			if afterBreak {
				r.discardByte()
			}
			break
		}

		r.discardByte()
	}

	return nil
}

var errSignificantInCont = fmt.Errorf("%w: unexpected significant token in escline", ErrInvalidSyntax)

// readUntilSignificant allows the provided reader to skip whitespace and comments.
//
// Note: this method will NOT skip over new lines.
func readUntilSignificant(r *reader, insideNode bool) error {

	escapedLine := false

outer:
	for {

		ch, err := r.peekRune()
		if err != nil {
			return err
		}

		if isWhitespace(ch) {
			r.discardBytes(utf8.RuneLen(ch))
			continue
		}

		// Check for line continuation
		if ch == '\\' && insideNode {
			r.discardByte()
			escapedLine = true
			continue
		}

		// Check for single-line comments
		if comment, err := r.isNext(charsStartComment[:]); comment && err == nil {
			r.discardBytes(2)
			return skipUntilNewLine(r, true)
		}

		// Check for multiline comments
		if comment, err := r.isNext(charsStartCommentBlock[:]); comment && err == nil {
			r.discardBytes(2)
			// Per spec, multiline comments can be nested, so we can't do naive ReadString("*/")
			depth := 1
		inner:
			for {

				start, err := r.isNext(charsStartCommentBlock[:])
				if err != nil {
					return err
				}

				if start {
					depth += 1
					r.discardBytes(2)
					continue inner
				}

				end, err := r.isNext(charsEndCommentBlock[:])
				if err != nil {
					return err
				}

				if end {
					r.discardBytes(2)
					depth -= 1
					if depth <= 0 {
						continue outer
					} else {
						continue inner
					}
				}

				r.discardByte()
			}
		}

		if escapedLine {
			if isNewLine(ch) {
				if err := skipUntilNewLine(r, true); err != nil {
					return err
				}
				escapedLine = false
				continue
			}
			return errSignificantInCont
		}

		return nil
	}
}
