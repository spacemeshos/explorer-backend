package rest

import (
    "bytes"
    "fmt"
    "unicode/utf8"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"

    "github.com/spacemeshos/explorer-backend/model"
)
/*
200:
{
  data: [
    {id: 1},
    {id: 2},
    {id: 3},
    {id: 4},
    ],
  pagination: {
    totalCount: 100,
    pageCount: 5,
    perPage: 20,
    hasNext: true,
    next: 2,
    hasPrevious: false,
    current: 1,
    previous: 1
  }
} 

error:

{
  error: {
    status: 404,
    message: 'Not Found',
  }
}
*/

var hex = "0123456789abcdef"

// safeSet holds the value true if the ASCII character with the given array
// position can be represented inside a JSON string without any further
// escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), and the backslash character ("\").
var safeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

// htmlSafeSet holds the value true if the ASCII character with the given
// array position can be safely represented inside a JSON string, embedded
// inside of HTML <script> tags, without any additional escaping.
//
// All values are true except for the ASCII control characters (0-31), the
// double quote ("), the backslash character ("\"), HTML opening and closing
// tags ("<" and ">"), and the ampersand ("&").
var htmlSafeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      false,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      false,
	'=':      true,
	'>':      false,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}
func encodeString(e *bytes.Buffer, s string, escapeHTML bool) {
    e.WriteByte('"')
    start := 0
    for i := 0; i < len(s); {
        if b := s[i]; b < utf8.RuneSelf {
            if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
                i++
                continue
            }
            if start < i {
                e.WriteString(s[start:i])
            }
            e.WriteByte('\\')
            switch b {
            case '\\', '"':
                e.WriteByte(b)
            case '\n':
                e.WriteByte('n')
            case '\r':
                e.WriteByte('r')
            case '\t':
                e.WriteByte('t')
            default:
                // This encodes bytes < 0x20 except for \t, \n and \r.
                // If escapeHTML is set, it also escapes <, >, and &
                // because they can lead to security holes when
                // user-controlled strings are rendered into JSON
                // and served to some browsers.
                e.WriteString(`u00`)
                e.WriteByte(hex[b>>4])
                e.WriteByte(hex[b&0xF])
            }
            i++
            start = i
            continue
        }
        c, size := utf8.DecodeRuneInString(s[i:])
        if c == utf8.RuneError && size == 1 {
            if start < i {
                e.WriteString(s[start:i])
            }
            e.WriteString(`\ufffd`)
            i += size
            start = i
            continue
        }
        // U+2028 is LINE SEPARATOR.
        // U+2029 is PARAGRAPH SEPARATOR.
        // They are both technically valid characters in JSON strings,
        // but don't work in JSONP, which has to be evaluated as JavaScript,
        // and can lead to security holes there. It is valid JSON to
        // escape them, so we do so unconditionally.
        // See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
        if c == '\u2028' || c == '\u2029' {
            if start < i {
                e.WriteString(s[start:i])
            }
            e.WriteString(`\u202`)
            e.WriteByte(hex[c&0xF])
            i += size
            start = i
            continue
        }
        i += size
    }

    if start < len(s) {
        e.WriteString(s[start:])
    }

    e.WriteByte('"')

}

func writeD(buf *bytes.Buffer, d *bson.D) {
    var needSeparator bool
    buf.WriteByte('{')
    for _, e := range *d {
//        fmt.Println(reflect.TypeOf(e.Value).String())
        if needSeparator {
            buf.WriteByte(',')
        } else {
            needSeparator = true
        }
        if inner, ok := e.Value.(bson.D); ok {
            buf.WriteByte('"')
            buf.WriteString(e.Key)
            buf.WriteString("\":")
            writeD(buf, &inner)
        } else if s, ok := e.Value.(string); ok {
            buf.WriteByte('"')
            buf.WriteString(e.Key)
            buf.WriteString("\":")
            encodeString(buf, s, true)
        } else if a, ok := e.Value.([]interface{}); ok {
            buf.WriteByte('"')
            buf.WriteString(e.Key)
            buf.WriteString("\":")
            writeA(buf, a)
        } else if id, ok := e.Value.(primitive.ObjectID); ok {
            buf.WriteByte('"')
            buf.WriteString(e.Key)
            buf.WriteString("\":")
            encodeString(buf, id.Hex(), true)
        } else {
            buf.WriteByte('"')
            buf.WriteString(e.Key)
            buf.WriteString("\":")
            buf.WriteString(fmt.Sprintf("%v", e.Value))
        }
    }
    buf.WriteByte('}')
}

func writeA(buf *bytes.Buffer, a []interface{}) {
    var needSeparator bool
    buf.WriteByte('[')
    if a != nil {
        for _, item := range a {
            if needSeparator {
                buf.WriteByte(',')
            } else {
                needSeparator = true
            }
            if d, ok := item.(bson.D); ok {
                writeD(buf, &d)
            } else if s, ok := item.(string); ok {
                buf.WriteString(s)
            } else if subA, ok := item.([]interface{}); ok {
                writeA(buf, subA)
            } else {
                buf.WriteString(fmt.Sprintf("%v", item))
            }
        }
    }
    buf.WriteByte(']')
}

func setDataInfo(buf *bytes.Buffer, data []bson.D) error {
    var needSeparator bool
    buf.WriteString("\"data\":[")
    if data != nil {
        for _, item := range data {
            if needSeparator {
                buf.WriteByte(',')
            } else {
                needSeparator = true
            }
            writeD(buf, &item)
        }
    }
    buf.WriteByte(']')
    return nil
}

func fixCheckedAddress(data []bson.D, positions []int) error {
    if data != nil {
        for i, _ := range data {
            for _, j := range positions {
                if address, ok := data[i][j].Value.(string); ok {
                    data[i][j].Value = model.ToCheckedAddress(address)
                }
            }
        }
    }
    return nil
}
