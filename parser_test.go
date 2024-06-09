package goson

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestParseNull(t *testing.T) {
	var p Parser
	v, err := p.Parse("null")
	assert.Nil(t, err)
	assert.Equal(t, NULL, v.getType())
}

func TestParseTrue(t *testing.T) {
	var p Parser
	v, err := p.Parse("true")
	assert.Nil(t, err)
	assert.Equal(t, TRUE, v.getType())
}

func TestParseFalse(t *testing.T) {
	var p Parser
	v, err := p.Parse("false")
	assert.Nil(t, err)
	assert.Equal(t, FALSE, v.getType())
}

func TestParseNumber(t *testing.T) {
	var p Parser
	f := func(n float64, s string) {
		v, err := p.Parse(s)
		assert.Nil(t, err)
		assert.Equal(t, NUMBER, v.getType())
		assert.Equal(t, n, v.n)
	}

	f(0.0, "0")
	f(0.0, "-0")
	f(0.0, "-0.0")
	f(1.0, "1")
	f(-1.0, "-1")
	f(1.5, "1.5")
	f(-1.5, "-1.5")
	f(3.1416, "3.1416")
	f(1e10, "1E10")
	f(1e10, "1e10")
	f(1e+10, "1E+10")
	f(1e-10, "1E-10")
	f(-1e10, "-1E10")
	f(-1e10, "-1e10")
	f(-1e+10, "-1E+10")
	f(-1e-10, "-1E-10")
	f(1.234e+10, "1.234E+10")
	f(1.234e-10, "1.234E-10")
	f(0.0, "1e-10000") /* must underflow */

	f(1.0000000000000002, "1.0000000000000002")           /* the smallest number > 1 */
	f(4.9406564584124654e-324, "4.9406564584124654e-324") /* minimum denormal */
	f(-4.9406564584124654e-324, "-4.9406564584124654e-324")
	f(2.2250738585072009e-308, "2.2250738585072009e-308") /* Max subnormal double */
	f(-2.2250738585072009e-308, "-2.2250738585072009e-308")
	f(2.2250738585072014e-308, "2.2250738585072014e-308") /* Min normal positive double */
	f(-2.2250738585072014e-308, "-2.2250738585072014e-308")
	f(1.7976931348623157e+308, "1.7976931348623157e+308") /* Max double */
	f(-1.7976931348623157e+308, "-1.7976931348623157e+308")
}

func TestParseString(t *testing.T) {
	var p Parser
	f := func(e, s string) {
		v, err := p.Parse(s)
		assert.Nil(t, err)
		assert.Equal(t, STRING, v.getType())
		str, err := v.getString()
		assert.Nil(t, err)
		assert.Equal(t, e, str)
	}

	f("", "\"\"")
	f("Hello", "\"Hello\"")
	f("Hello\nWorld", "\"Hello\\nWorld\"")
	f("\" \\ / \b \f \n \r \t", "\"\\\" \\\\ \\/ \\b \\f \\n \\r \\t\"")
	f("Hello\000World", "\"Hello\\u0000World\"")
	f("\x24", "\"\\u0024\"")                    /* Dollar sign U+0024 */
	f("\xC2\xA2", "\"\\u00A2\"")                /* Cents sign U+00A2 */
	f("\xE2\x82\xAC", "\"\\u20AC\"")            /* Euro sign U+20AC */
	f("\xF0\x9D\x84\x9E", "\"\\uD834\\uDD1E\"") /* G clef sign U+1D11E */
	f("\xF0\x9D\x84\x9E", "\"\\ud834\\udd1e\"") /* G clef sign U+1D11E */
}

func TestParseArray(t *testing.T) {
	var p Parser

	v, err := p.Parse("[ ]")
	assert.Nil(t, err)
	assert.Equal(t, ARRAY, v.getType())
	assert.Equal(t, 0, len(v.a))

	v, err = p.Parse("[ null , false , true, 123 , \"abc\" ]")
	assert.Nil(t, err)
	assert.Equal(t, ARRAY, v.getType())
	assert.Equal(t, 5, len(v.a))

	e, err := v.getArrayElement(0)
	assert.Nil(t, err)
	assert.Equal(t, NULL, e.getType())

	e, err = v.getArrayElement(1)
	assert.Nil(t, err)
	assert.Equal(t, FALSE, e.getType())

	e, err = v.getArrayElement(2)
	assert.Nil(t, err)
	assert.Equal(t, TRUE, e.getType())

	e, err = v.getArrayElement(3)
	assert.Nil(t, err)
	assert.Equal(t, NUMBER, e.getType())
	f, err := e.getNumber()
	assert.Nil(t, err)
	assert.Equal(t, 123.0, f)

	e, err = v.getArrayElement(4)
	assert.Nil(t, err)
	assert.Equal(t, STRING, e.getType())
	s, err := e.getString()
	assert.Nil(t, err)
	assert.Equal(t, "abc", s)

	v, err = p.Parse("[ [ ] , [ 0 ] , [ 0, 1 ] , [ 0, 1, 2 ] ]")
	assert.Nil(t, err)
	assert.Equal(t, ARRAY, v.getType())
	assert.Equal(t, 4, len(v.a))
	for i := 0; i < 4; i++ {
		e, err = v.getArrayElement(i)
		assert.Nil(t, err)
		assert.Equal(t, ARRAY, e.getType())
		assert.Equal(t, i, len(e.a))
		for j := 0; j < i; j++ {
			ee, err := e.getArrayElement(j)
			assert.Nil(t, err)
			assert.Equal(t, NUMBER, ee.getType())
			f, err := ee.getNumber()
			assert.Nil(t, err)
			assert.Equal(t, float64(j), f)
		}
	}
}

func TestParseObject(t *testing.T) {
	var p Parser

	v, err := p.Parse("{ }")
	assert.Nil(t, err)
	assert.Equal(t, OBJECT, v.getType())
	assert.Equal(t, 0, len(v.o))

	var s string
	s += " { "
	s += "\"n\" : null , "
	s += "\"f\" : false , "
	s += "\"t\" : true , "
	s += "\"i\" : 123 , "
	s += "\"s\" : \"abc\", "
	s += "\"a\" : [ 1, 2, 3 ],"
	s += "\"o\" : { \"1\" : 1, \"2\" : 2, \"3\" : 3 }"
	s += " } "
	v, err = p.Parse(s)
	assert.Nil(t, err)
	assert.Equal(t, OBJECT, v.getType())
	assert.Equal(t, 7, len(v.o))

	vv, err := v.getObjectValue("n")
	assert.Nil(t, err)
	assert.Equal(t, NULL, vv.getType())

	vv, err = v.getObjectValue("f")
	assert.Nil(t, err)
	assert.Equal(t, FALSE, vv.getType())

	vv, err = v.getObjectValue("t")
	assert.Nil(t, err)
	assert.Equal(t, TRUE, vv.getType())

	vv, err = v.getObjectValue("i")
	assert.Nil(t, err)
	assert.Equal(t, NUMBER, vv.getType())
	f, err := vv.getNumber()
	assert.Nil(t, err)
	assert.Equal(t, 123.0, f)

	vv, err = v.getObjectValue("s")
	assert.Nil(t, err)
	assert.Equal(t, STRING, vv.getType())
	ss, err := vv.getString()
	assert.Nil(t, err)
	assert.Equal(t, "abc", ss)

	vv, err = v.getObjectValue("a")
	assert.Nil(t, err)
	assert.Equal(t, ARRAY, vv.getType())
	assert.Equal(t, 3, len(vv.a))
	for i := 0; i < 3; i++ {
		e, err := vv.getArrayElement(i)
		assert.Nil(t, err)
		assert.Equal(t, NUMBER, e.getType())
		f, err = e.getNumber()
		assert.Nil(t, err)
		assert.Equal(t, float64(i+1), f)
	}

	vv, err = v.getObjectValue("o")
	assert.Nil(t, err)
	assert.Equal(t, OBJECT, vv.getType())
	assert.Equal(t, 3, len(vv.o))
	for i := 0; i < 3; i++ {
		vvv, err := vv.getObjectValue(strconv.Itoa(i + 1))
		assert.Nil(t, err)
		assert.Equal(t, NUMBER, vvv.getType())
		f, err = vvv.getNumber()
		assert.Nil(t, err)
		assert.Equal(t, float64(i+1), f)
	}
}

func parseError(t *testing.T, e error, s string) {
	var p Parser
	v, err := p.Parse(s)
	assert.Equal(t, e, err)
	assert.Equal(t, NULL, v.getType())
}

func TestParseExpectValue(t *testing.T) {
	parseError(t, nil, "")
	parseError(t, nil, " ")
}

func TestParseInvalidValue(t *testing.T) {
	parseError(t, ErrParseInvalidValue, "nul")
	parseError(t, ErrParseInvalidValue, "?")
	parseError(t, ErrParseInvalidValue, "+0")
	parseError(t, ErrParseInvalidValue, "+1")
	parseError(t, ErrParseInvalidValue, ".123") /* at least one digit before '.' */
	parseError(t, ErrParseInvalidValue, "1.")   /* at least one digit after '.' */
	parseError(t, ErrParseInvalidValue, "1e")
	parseError(t, ErrParseInvalidValue, "INF")
	parseError(t, ErrParseInvalidValue, "inf")
	parseError(t, ErrParseInvalidValue, "NAN")
	parseError(t, ErrParseInvalidValue, "nan")
	parseError(t, ErrParseInvalidValue, "[1,]")
	parseError(t, ErrParseInvalidValue, "[\"a\", nul]")
}

func TestParseRootNotSingular(t *testing.T) {
	parseError(t, ErrParseRootNotSingular, "null x")
	parseError(t, ErrParseRootNotSingular, "0123")
	parseError(t, ErrParseRootNotSingular, "0x0")
	parseError(t, ErrParseRootNotSingular, "0x123")
}

func TestParseNumberTooBig(t *testing.T) {
	parseError(t, ErrParseNumberTooBig, "1e309")
	parseError(t, ErrParseNumberTooBig, "-1e309")
}

func TestParseMissingQuotationMark(t *testing.T) {
	parseError(t, ErrParseMissQuotationMark, "\"")
	parseError(t, ErrParseMissQuotationMark, "\"abc")
}

func TestParseInvalidStringEscape(t *testing.T) {
	parseError(t, ErrParseInvalidStringEscape, "\"\\v\"")
	parseError(t, ErrParseInvalidStringEscape, "\"\\'\"")
	parseError(t, ErrParseInvalidStringEscape, "\"\\0\"")
	parseError(t, ErrParseInvalidStringEscape, "\"\\x12\"")
}

func TestParseInvalidStringChar(t *testing.T) {
	parseError(t, ErrParseInvalidStringChar, "\"\x01\"")
	parseError(t, ErrParseInvalidStringChar, "\"\x1F\"")
}

func TestParseInvalidUnicodeHex(t *testing.T) {
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u0\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u01\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u012\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u/000\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\uG000\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u0/00\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u0G00\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u00/0\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u00G0\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u000/\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u000G\"")
	parseError(t, ErrParseInvalidUnicodeHex, "\"\\u 123\"")
}

func TestParseInvalidUnicodeSurrogate(t *testing.T) {
	parseError(t, ErrParseInvalidUnicodeSurrogate, "\"\\uD800\"")
	parseError(t, ErrParseInvalidUnicodeSurrogate, "\"\\uDBFF\"")
	parseError(t, ErrParseInvalidUnicodeSurrogate, "\"\\uD800\\\\\"")
	parseError(t, ErrParseInvalidUnicodeSurrogate, "\"\\uD800\\uDBFF\"")
	parseError(t, ErrParseInvalidUnicodeSurrogate, "\"\\uD800\\uE000\"")
}

func TestParseMissCommaOrSquareBracket(t *testing.T) {
	parseError(t, ErrParseMissCommaOrSquareBracket, "[1")
	parseError(t, ErrParseMissCommaOrSquareBracket, "[1}")
	parseError(t, ErrParseMissCommaOrSquareBracket, "[1 2")
	parseError(t, ErrParseMissCommaOrSquareBracket, "[[]")
}

func TestParseMissKey(t *testing.T) {
	parseError(t, ErrParseMissKey, "{:1,")
	parseError(t, ErrParseMissKey, "{1:1,")
	parseError(t, ErrParseMissKey, "{true:1,")
	parseError(t, ErrParseMissKey, "{false:1,")
	parseError(t, ErrParseMissKey, "{null:1,")
	parseError(t, ErrParseMissKey, "{[]:1,")
	parseError(t, ErrParseMissKey, "{{}:1,")
	parseError(t, ErrParseMissKey, "{\"a\":1,")
}

func TestParseMissColon(t *testing.T) {
	parseError(t, ErrParseMissColon, "{\"a\"}")
	parseError(t, ErrParseMissColon, "{\"a\",\"b\"}")
}

func TestParseMissCommaOrCurlyBracket(t *testing.T) {
	parseError(t, ErrParseMissCommaOrCurlyBracket, "{\"a\":1")
	parseError(t, ErrParseMissCommaOrCurlyBracket, "{\"a\":1]")
	parseError(t, ErrParseMissCommaOrCurlyBracket, "{\"a\":1 \"b\"")
	parseError(t, ErrParseMissCommaOrCurlyBracket, "{\"a\":{}")
}

func parseRoundTrip(t *testing.T, s string) {
	var p Parser
	v, err := p.Parse(s)
	assert.Nil(t, err)
	ss := v.stringifyValue()
	assert.Equal(t, s, ss)
}

func TestStringifyBasic(t *testing.T) {
	parseRoundTrip(t, "null")
	parseRoundTrip(t, "false")
	parseRoundTrip(t, "true")
}

func TestStringifyNumber(t *testing.T) {
	parseRoundTrip(t, "0")
	parseRoundTrip(t, "-0")
	parseRoundTrip(t, "1")
	parseRoundTrip(t, "-1")
	parseRoundTrip(t, "1.5")
	parseRoundTrip(t, "-1.5")
	parseRoundTrip(t, "3.25")
	parseRoundTrip(t, "1e+20")
	parseRoundTrip(t, "1.234e+20")
	parseRoundTrip(t, "1.234e-20")

	parseRoundTrip(t, "1.0000000000000002")      /* the smallest number > 1 */
	parseRoundTrip(t, "4.9406564584124654e-324") /* minimum denormal */
	parseRoundTrip(t, "-4.9406564584124654e-324")
	parseRoundTrip(t, "2.2250738585072009e-308") /* Max subnormal double */
	parseRoundTrip(t, "-2.2250738585072009e-308")
	parseRoundTrip(t, "2.2250738585072014e-308") /* Min normal positive double */
	parseRoundTrip(t, "-2.2250738585072014e-308")
	parseRoundTrip(t, "1.7976931348623157e+308") /* Max double */
	parseRoundTrip(t, "-1.7976931348623157e+308")
}

func TestStringifyString(t *testing.T) {
	parseRoundTrip(t, "\"\"")
	parseRoundTrip(t, "\"Hello\"")
	parseRoundTrip(t, "\"Hello\\nWorld\"")
	parseRoundTrip(t, "\"\\\" \\\\ / \\b \\f \\n \\r \\t\"")
	parseRoundTrip(t, "\"Hello\\u0000World\"")
}

func TestStringifyArray(t *testing.T) {
	parseRoundTrip(t, "[]")
	parseRoundTrip(t, "[null,false,true,123,\"abc\",[1,2,3]]")
}

func TestStringifyObject(t *testing.T) {
	parseRoundTrip(t, "{}")
	parseRoundTrip(t, "{\"n\":null,\"f\":false,\"t\":true,\"i\":123,\"s\":\"abc\",\"a\":[1,2,3],\"o\":{\"1\":1,\"2\":2,\"3\":3}}")
}

func TestAccessNull(t *testing.T) {
	var v Value
	v.setNull()
	assert.Equal(t, NULL, v.getType())
}

func TestAccessBoolean(t *testing.T) {
	var v Value
	v.setBoolean(true)
	b, err := v.getBoolean()
	assert.Nil(t, err)
	assert.Equal(t, true, b)
	v.setBoolean(false)
	b, err = v.getBoolean()
	assert.Nil(t, err)
	assert.Equal(t, false, b)
}

func TestAccessNumber(t *testing.T) {
	var v Value
	v.setNumber(1234.5)
	n, err := v.getNumber()
	assert.Nil(t, err)
	assert.Equal(t, 1234.5, n)
}

func TestAccessString(t *testing.T) {
	var v Value
	v.setString("")
	s, err := v.getString()
	assert.Nil(t, err)
	assert.Equal(t, "", s)
	v.setString("Hello")
	s, err = v.getString()
	assert.Nil(t, err)
	assert.Equal(t, "Hello", s)
}

func TestAccessArray(t *testing.T) {
	var v Value

	for i := 0; i <= 5; i += 5 {
		v.setArray(i)
		assert.Equal(t, ARRAY, v.getType())
		assert.Equal(t, 0, len(v.a))
		assert.Equal(t, i, cap(v.a))
		for j := 0; j < 10; j++ {
			var e Value
			e.setNumber(float64(j))
			_ = v.insertArrayElement(&e, j)
		}
		assert.Equal(t, 10, len(v.a))
		for j := 0; j < 10; j++ {
			e, err := v.getArrayElement(j)
			assert.Nil(t, err)
			f, err := e.getNumber()
			assert.Nil(t, err)
			assert.Equal(t, float64(j), f)
		}
	}

	err := v.eraseArrayElement(len(v.a)-1, 1)
	assert.Nil(t, err)
	assert.Equal(t, 9, len(v.a))
	for j := 0; j < 9; j++ {
		e, err := v.getArrayElement(j)
		assert.Nil(t, err)
		f, err := e.getNumber()
		assert.Nil(t, err)
		assert.Equal(t, float64(j), f)
	}

	err = v.eraseArrayElement(4, 0)
	assert.Nil(t, err)
	assert.Equal(t, 9, len(v.a))
	for j := 0; j < 9; j++ {
		e, err := v.getArrayElement(j)
		assert.Nil(t, err)
		f, err := e.getNumber()
		assert.Nil(t, err)
		assert.Equal(t, float64(j), f)
	}

	err = v.eraseArrayElement(8, 1)
	assert.Nil(t, err)
	assert.Equal(t, 8, len(v.a))
	for j := 0; j < 8; j++ {
		e, err := v.getArrayElement(j)
		assert.Nil(t, err)
		f, err := e.getNumber()
		assert.Nil(t, err)
		assert.Equal(t, float64(j), f)
	}

	err = v.eraseArrayElement(0, 2)
	assert.Nil(t, err)
	assert.Equal(t, 6, len(v.a))
	for j := 0; j < 6; j++ {
		e, err := v.getArrayElement(j)
		assert.Nil(t, err)
		f, err := e.getNumber()
		assert.Nil(t, err)
		assert.Equal(t, float64(j+2), f)
	}

	for i := 0; i < 2; i++ {
		var e Value
		e.setNumber(float64(i))
		err = v.insertArrayElement(&e, i)
		assert.Nil(t, err)
	}
	assert.Equal(t, 8, len(v.a))
	for j := 0; j < 8; j++ {
		e, err := v.getArrayElement(j)
		assert.Nil(t, err)
		f, err := e.getNumber()
		assert.Nil(t, err)
		assert.Equal(t, float64(j), f)
	}

	err = v.clearArray()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(v.a))
}

func TestAccessObject(t *testing.T) {
	var v Value

	for i := 0; i <= 5; i += 5 {
		v.setObject(i)
		assert.Equal(t, OBJECT, v.getType())
		assert.Equal(t, 0, len(v.o))
		assert.Equal(t, i, cap(v.o))
		for j := 0; j < 10; j++ {
			k := string(byte('a' + j))
			var value Value
			value.setNumber(float64(j))
			err := v.setObjectValue(k, &value)
			assert.Nil(t, err)
		}
		for j := 0; j < 10; j++ {
			k := string(byte('a' + j))
			value, err := v.getObjectValue(k)
			assert.Nil(t, err)
			f, err := value.getNumber()
			assert.Nil(t, err)
			assert.Equal(t, float64(j), f)
		}
	}

	_, err := v.getObjectValue("j")
	assert.Nil(t, err)
	err = v.removeObjectValue("j")
	assert.Nil(t, err)
	_, err = v.getObjectValue("j")
	assert.Equal(t, ErrKeyNotExist, err)
	assert.Equal(t, 9, len(v.o))

	_, err = v.getObjectValue("a")
	assert.Nil(t, err)
	err = v.removeObjectValue("a")
	assert.Nil(t, err)
	_, err = v.getObjectValue("a")
	assert.Equal(t, ErrKeyNotExist, err)
	assert.Equal(t, 8, len(v.o))

	for i := 0; i < 8; i++ {
		k := string(byte('a' + i + 1))
		vv, err := v.getObjectValue(k)
		assert.Nil(t, err)
		f, err := vv.getNumber()
		assert.Nil(t, err)
		assert.Equal(t, float64(i+1), f)
	}

	var value Value
	value.setString("World")
	err = v.setObjectValue("Hello", &value)
	assert.Nil(t, err)
	vv, err := v.getObjectValue("Hello")
	assert.Nil(t, err)
	assert.Equal(t, "World", vv.s)

	err = v.clearObject()
	assert.Nil(t, err)
	assert.Equal(t, 0, len(v.o))
}

func TestIsEqual(t *testing.T) {
	f := func(t *testing.T, l, r string, b bool) {
		var p Parser
		lv, err := p.Parse(l)
		assert.Nil(t, err)
		rv, err := p.Parse(r)
		assert.Nil(t, err)
		assert.Equal(t, b, isEqual(lv, rv))
	}
	f(t, "true", "true", true)
	f(t, "true", "false", false)
	f(t, "false", "false", true)
	f(t, "null", "null", true)
	f(t, "null", "0", false)
	f(t, "123", "123", true)
	f(t, "123", "456", false)
	f(t, "\"abc\"", "\"abc\"", true)
	f(t, "\"abc\"", "\"abcd\"", false)
	f(t, "[]", "[]", true)
	f(t, "[]", "null", false)
	f(t, "[1,2,3]", "[1,2,3]", true)
	f(t, "[1,2,3]", "[1,2,3,4]", false)
	f(t, "[[]]", "[[]]", true)
	f(t, "{}", "{}", true)
	f(t, "{}", "null", false)
	f(t, "{}", "[]", false)
	f(t, "{\"a\":1,\"b\":2}", "{\"a\":1,\"b\":2}", true)
	f(t, "{\"a\":1,\"b\":2}", "{\"b\":2,\"a\":1}", true)
	f(t, "{\"a\":1,\"b\":2}", "{\"a\":1,\"b\":3}", false)
	f(t, "{\"a\":1,\"b\":2}", "{\"a\":1,\"b\":2,\"c\":3}", false)
	f(t, "{\"a\":{\"b\":{\"c\":{}}}}", "{\"a\":{\"b\":{\"c\":{}}}}", true)
	f(t, "{\"a\":{\"b\":{\"c\":{}}}}", "{\"a\":{\"b\":{\"c\":[]}}}", false)
}

func TestCopy(t *testing.T) {
	var p Parser
	v1, err := p.Parse("{\"t\":true,\"f\":false,\"n\":null,\"d\":1.5,\"a\":[1,2,3]}")
	assert.Nil(t, err)
	v2 := v1.copy()
	assert.True(t, isEqual(v1, v2))
}
