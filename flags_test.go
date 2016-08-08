// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package flags

import (
	"flag"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type logging struct {
	Interval int
	Path     string
}

type socket struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type tcp struct {
	ReadTimeout time.Duration
	socket
}

type network struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	tcp
}

type Cfg1 struct {
	logging
	network
}

func TestFlagMakerExample(t *testing.T) {
	cfg := Cfg1{}

	args := []string{
		"--network.tcp.socket.readtimeout", "5ms",
		"--network.tcp.readtimeout", "3ms",
		"-logging.path", "/var/log",
	}
	args, err := ParseArgs(cfg, args)
	assert.False(t, err == nil)
	args, err = ParseArgs(&cfg, args)
	assert.True(t, err == nil)
	assert.Equal(t, 0, len(args))

	expected := Cfg1{
		network: network{
			tcp: tcp{
				ReadTimeout: time.Duration(3) * time.Millisecond,
				socket: socket{
					ReadTimeout: time.Duration(5) * time.Millisecond,
				},
			},
		},
		logging: logging{
			Path: "/var/log",
		},
	}
	assert.Equal(t, expected, cfg)
}

type auth struct {
	Token string
	Tag   float64
}

type credentials struct {
	User     string
	Password string
	auth
}

type database struct {
	DBName    string
	TableName string
	credentials
}

type Cfg2 struct {
	logging
	database
	*string
}

func TestFlagMakerExampleFlattened(t *testing.T) {
	cfg := Cfg2{}

	args := []string{
		"--dbname", "db1",
		"--token", "abcd",
		"-tag=3.14",
		"-path", "/var/log",
	}

	fm := NewFlagMakerAdv(&FlagMakingOptions{true, true, "not-care"})
	args, err := fm.ParseArgs(&cfg, args)

	assert.True(t, err == nil)
	assert.Equal(t, 0, len(args))

	expected := Cfg2{}
	expected.Tag = 3.14
	expected.DBName = "db1"
	expected.Token = "abcd"
	expected.Path = "/var/log"

	assert.Equal(t, expected, cfg)
}

type C4 struct {
	TableName string
}

type C3 struct {
	DBName string
	C4
}

type C2 struct {
	User     string
	Password int64
	Tag      int8
	C3
}

type C1 struct {
	Name       string `yaml:"label"`
	Value      int
	Float      float64
	Timeout    time.Duration
	Hosts      []string
	Ports      []int
	Weights    []float64
	Credential C2

	// some unexported fields
	opentimeout time.Duration
	localhost   string
}

func TestFlagMakerBasic(t *testing.T) {
	c := &C1{
		Name:    "basic",
		Value:   10,
		Float:   7.4,
		Timeout: time.Duration(10) * time.Millisecond,
		Hosts:   []string{"host1", "host2"},
		Ports:   []int{89, 90},
		Credential: C2{
			User:     "user",
			Password: 1234,
			Tag:      20,
			C3: C3{
				DBName: "db1",
				C4: C4{
					TableName: "t1",
				},
			},
		},
		opentimeout: time.Duration(3) * time.Microsecond,
		localhost:   "weird.host",
	}
	args := []string{
		"--label", "advanced", "-float", "5.1", "-timeout", "5ms", "--ports", "22", "--ports", "43",
		"--credential.user", "uber", "--credential.tag", "80", "--credential.c3.dbname", "db2",
		"--credential.c3.c4.tablename", "t2"}
	args, err := ParseArgs(c, args)
	assert.Equal(t, nil, err, "should be no error")
	assert.Equal(t, 0, len(args), "should be no arg left")

	expected := *c
	// only these fields should be modified by parsing the arguments.
	expected.Credential.User = "uber"
	expected.Credential.Password = int64(1234)
	expected.Credential.Tag = int8(80)
	expected.Name = "advanced"
	expected.Credential.C3.DBName = "db2"
	expected.Credential.C3.TableName = "t2"
	expected.Ports = []int{22, 43}

	assert.Equal(t, &expected, c)
}

type CTypes struct {
	Strval  string
	Bval    bool
	F32val  float32
	F64val  float64
	Ival    int
	I8val   int8
	I16val  int16
	I32val  int32
	I64val  int64
	UIval   uint
	UI8val  uint8
	UI16val uint16
	UI32val uint32
	UI64val uint64
}

func TestFlagMakerTypes(t *testing.T) {
	/* Check all of the types */
	refCtypes := &CTypes{
		Strval: "string value",
		Bval:   true,
		F32val: 3.1415927,         // <- Max PI for 32 bit float
		F64val: 3.141592653589793, // <- Max PI for 64 bit float
		/* The rest of these use the highest value for the type */
		Ival:    int(0x7fffffffffffffff),
		I8val:   int8(0x7f),
		I16val:  int16(0x7fff),
		I32val:  int32(0x7fffffff),
		I64val:  int64(0x7fffffffffffffff),
		UIval:   uint(0xffffffffffffffff),
		UI8val:  uint8(0xff),
		UI16val: uint16(0xffff),
		UI32val: uint32(0xffffffff),
		UI64val: uint64(0xffffffffffffffff),
	}
	parseCtypes := &CTypes{}
	args := []string{
		"-strval", "string value", "--bval",
		"-f32val", "3.1415927", "--f64val", "3.141592653589793",
		"--ival", "9223372036854775807", "--i8val", "127", "--i16val", "32767",
		"-i32val", "2147483647", "--i64val", "9223372036854775807",
		"--uival", "18446744073709551615", "--ui8val", "255", "--ui16val", "65535",
		"-ui32val", "4294967295", "--ui64val", "18446744073709551615"}
	args, err := ParseArgs(parseCtypes, args)
	assert.Equal(t, nil, err, "should be no error")
	assert.Equal(t, parseCtypes, refCtypes)
}

type D1 struct {
	F1 ****string
	F2 ***[]int
}

type D2 struct {
	F1 **[]float64
	F2 *****bool
	D1
}

type D3 struct {
	D2
	F3    uint
	F4    int64
	Hosts []string
}

type DD struct {
	D1
	D2
	D3
}

func TestFlagMakerComplex(t *testing.T) {
	d := DD{}
	args := []string{"-d2.f1", "1.2", "-d3.d2.d1.f2", "45", "-d2.f2", "-d2.f1", "4.2", "-d3.d2.d1.f2", "56", "-d2.f1", "7.4", "-d3.d2.d1.f2", "78"}
	args, err := ParseArgs(&d, args)
	assert.Equal(t, nil, err, "unexpected error")
	assert.Equal(t, 0, len(args))
	assert.Equal(t, true, *****d.D2.F2)
	assert.Equal(t, []int{45, 56, 78}, ***d.D3.D2.D1.F2)
	assert.Equal(t, []float64{1.2, 4.2, 7.4}, **d.D2.F1)
}

type I1 interface {
	Method1() string
}

type S1 struct {
	Host    string
	ignore  int
	Weights []float64
	F       int8
}

func (s *S1) Method1() string {
	return s.Host
}

type S2 struct {
	Open   bool
	Volume float64
}

func (s S2) Method1() string { return "haha" }

func TestFlagMakerInterface(t *testing.T) {
	var s I1 = &S1{ignore: 12}
	args := []string{
		"-host", "test.local", "-f", "16",
	}
	args, err := ParseArgs(s, args)
	assert.Equal(t, nil, err)
	expected := S1{
		Host:   "test.local",
		F:      int8(16),
		ignore: 12,
	}
	assert.Equal(t, "test.local", s.Method1())
	assert.Equal(t, &expected, s)
}

func TestFlagMakerPtrToIntf(t *testing.T) {
	s := &S1{}
	var i2 I1 = s
	args := []string{"--weights", "9.3", "--host", "www", "--weights", "10.0"}
	out, err := ParseArgs(&i2, args)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(out))
	assert.Equal(t, "www", s.Host)
	assert.Equal(t, []float64{9.3, 10.0}, s.Weights)

	s2 := S2{}
	var i3 I1 = s2
	args = []string{"--open", "--volume", "9.3"}
	out, err = ParseArgs(&i3, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interface must have pointer underlying type.")
	assert.Equal(t, 3, len(out))
}

type Cfg3 struct {
	D3
	*D2
}

func TestFlagMakerNested(t *testing.T) {
	cfg := Cfg3{}
	args := []string{"-d3.hosts", "h1.com", "-d2.f1", "1.2", "-d3.hosts", "h2.com", "-d2.f1", "4.2", "-d3.hosts", "h3.com", "-d2.f1", "7.4"}
	args, err := ParseArgs(&cfg, args)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(args))

	assert.Equal(t, []string{"h1.com", "h2.com", "h3.com"}, cfg.Hosts)
	assert.Equal(t, []float64{1.2, 4.2, 7.4}, **cfg.F1)
}

type Cfg4 struct {
	Name *string
	*string
	int
}

func TestFlagMakerUnnamedFields(t *testing.T) {
	c := Cfg4{int: 4}
	args := []string{"--name=haha"}
	args, err := ParseArgs(&c, args)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(args))
	ss := "haha"
	expected := Cfg4{Name: &ss, int: 4}
	assert.Equal(t, expected, c)
}

// The following test ensures that we can properly create flags for user
// defined non-struct types. The kind of an object and the type of an
// object is different. See the comments of defineFlag().
type String string
type Int int
type Int8 int8
type Int32 int32
type Int16 int16
type Int64 int64
type Float float64
type Float32 float32
type Uint uint
type Uint8 uint8
type Uint16 uint16
type Uint32 uint32
type Uint64 uint64
type Bool bool
type PString *String
type PInt *Int
type PInt8 *Int8
type PInt16 *Int16
type PInt32 *Int32
type PInt64 *Int64
type PFloat *Float
type PUint *Uint
type PUint8 *Uint8
type PUint16 *Uint16
type PUint32 *Uint32
type PUint64 *Uint64
type PBool *bool

type Cfg5 struct {
	S    String
	PS   PString
	I    Int
	PI   PInt
	I8   Int8
	PI8  PInt8
	I16  Int16
	PI16 PInt16
	I32  Int32
	PI32 PInt32
	I64  Int64
	PI64 PInt64
	F    Float
	PF   PFloat
	U    Uint
	U8   Uint8
	PU8  PUint8
	U16  Uint16
	PU16 PUint16
	U32  Uint32
	PU32 PUint32
	PU   PUint
	U64  Uint64
	PU64 PUint64
	B    Bool
	PB   PBool
	F32  Float32
	PF32 *Float32
}

func TestFlagMakerTypeDef(t *testing.T) {
	cfg := &Cfg5{}
	args := []string{"--s", "hehe", "--ps", "good",
		"-i", "33", "--pi", "44",
		"--i64", "55", "--pi64", "66",
		"--f", "5.7", "--pf", "6.7",
		"--u", "10", "--pu", "20",
		"--u64", "30", "--pu64", "40",
		"--i8", "20", "--pi8", "10",
		"--i16", "400", "-pi16", "500",
		"--i32", "600", "-pi32", "700",
		"--u8", "20", "--pu8", "10",
		"--u16", "400", "-pu16", "500",
		"--u32", "600", "-pu32", "700",
		"--f32", "10.1", "-pf32", "11.1",
		"--b=true", "--pb"}
	args, err := ParseArgs(cfg, args)

	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(args))

	assert.Equal(t, "hehe", string(cfg.S))
	assert.Equal(t, "good", string(*cfg.PS))

	assert.Equal(t, 33, int(cfg.I))
	assert.Equal(t, 44, int(*cfg.PI))

	assert.Equal(t, int8(20), int8(cfg.I8))
	assert.Equal(t, int8(10), int8(*cfg.PI8))

	assert.Equal(t, int16(400), int16(cfg.I16))
	assert.Equal(t, int16(500), int16(*cfg.PI16))

	assert.Equal(t, int32(600), int32(cfg.I32))
	assert.Equal(t, int32(700), int32(*cfg.PI32))

	assert.Equal(t, int64(55), int64(cfg.I64))
	assert.Equal(t, int64(66), int64(*cfg.PI64))

	assert.Equal(t, 5.7, float64(cfg.F))
	assert.Equal(t, 6.7, float64(*cfg.PF))

	assert.Equal(t, float32(10.1), float32(cfg.F32))
	assert.Equal(t, float32(11.1), float32(*cfg.PF32))

	assert.Equal(t, uint(10), uint(cfg.U))
	assert.Equal(t, uint(20), uint(*cfg.PU))

	assert.Equal(t, uint8(20), uint8(cfg.U8))
	assert.Equal(t, uint8(10), uint8(*cfg.PU8))

	assert.Equal(t, uint16(400), uint16(cfg.U16))
	assert.Equal(t, uint16(500), uint16(*cfg.PU16))

	assert.Equal(t, uint32(600), uint32(cfg.U32))
	assert.Equal(t, uint32(700), uint32(*cfg.PU32))

	assert.Equal(t, uint64(30), uint64(cfg.U64))
	assert.Equal(t, uint64(40), uint64(*cfg.PU64))

	assert.Equal(t, true, bool(cfg.B))
	assert.Equal(t, true, bool(*cfg.PB))
}

func TestFlagMakerInvalidInput(t *testing.T) {
	var cfg *Cfg5
	args := []string{"--s", "hehe", "--ps", "good",
		"-i", "33", "--pi", "44",
		"--i64", "55", "--pi64", "66",
	}
	out, err := ParseArgs(cfg, args)
	assert.Error(t, err)
	assert.Equal(t, "top level object cannot be nil", err.Error())
	assert.Equal(t, len(args), len(out))

	var cfg2 Cfg4
	out, err = ParseArgs(cfg2, args)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "top level object must be a pointer")
	assert.Equal(t, len(args), len(out))
}

func TestFlagMakerUnsupportedTypes(t *testing.T) {
	cases := []struct {
		cfg  interface{}
		args []string
	}{
		{&struct {
			Env   map[string]string
			Level int
		}{}, []string{"--level", "10", "--env", "hh,fgg,10"}},
		{&struct {
			Env   chan int
			Level int
		}{}, []string{"--level", "10", "--env", "hh,fgg,10"}},
		{&struct {
			Env   func(int) string
			Level int
		}{}, []string{"--level", "10", "--env", "hh,fgg,10"}},
	}

	for _, c := range cases {
		out, err := ParseArgs(c.cfg, c.args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "flag provided but not defined")
		assert.Equal(t, 1, len(out))
	}
}

func TestFlagMakerInvalidValue(t *testing.T) {
	cases := []struct {
		cfg  interface{}
		args []string
	}{
		{&struct{ Level int }{}, []string{"--level", "haha"}},
		{&struct{ Level int8 }{}, []string{"--level", "haha"}},
		{&struct{ Level int16 }{}, []string{"--level", "haha"}},
		{&struct{ Level int32 }{}, []string{"--level", "haha"}},
		{&struct{ Level int64 }{}, []string{"--level", "haha"}},
		{&struct{ Level uint8 }{}, []string{"--level", "haha"}},
		{&struct{ Level uint16 }{}, []string{"--level", "haha"}},
		{&struct{ Level uint32 }{}, []string{"--level", "haha"}},
		{&struct{ Level uint64 }{}, []string{"--level", "haha"}},
		{&struct{ Level float32 }{}, []string{"--level", "haha"}},
		{&struct{ Level float64 }{}, []string{"--level", "haha"}},
	}

	for _, c := range cases {
		out, err := ParseArgs(c.cfg, c.args)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value")
		// args are consumed even thought the value is invalid
		assert.Equal(t, 0, len(out))
	}
}

// slice

func TestFlagMakerStringSlice(t *testing.T) {
	type C struct {
		Hosts []string
	}
	cases := []struct {
		cfg            *C
		args, expected []string
	}{
		{&C{}, []string{"--hosts", "h1", "--hosts", "h2", "--hosts", "h3"}, []string{"h1", "h2", "h3"}},
		{&C{[]string{}}, []string{"--hosts", "h1", "--hosts", "h2", "--hosts", "h3"}, []string{"h1", "h2", "h3"}},
		{&C{}, []string{}, nil},
		{&C{[]string{}}, []string{}, []string{}},
		{&C{[]string{"l1", "l2"}}, []string{}, []string{"l1", "l2"}},
		{&C{[]string{"l1", "l2"}}, []string{"--hosts", "ok"}, []string{"ok"}},
	}
	for _, c := range cases {
		args, err := ParseArgs(c.cfg, c.args)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(args))
		assert.Equal(t, c.expected, c.cfg.Hosts)
	}
}

func TestFlagMakerIntSlice(t *testing.T) {
	type C struct {
		Levels []int
	}

	cases := []struct {
		cfg      *C
		args     []string
		expected []int
	}{
		{&C{}, []string{"--levels", "8", "--levels", "9", "--levels", "10"}, []int{8, 9, 10}},
		{&C{[]int{}}, []string{"--levels", "8", "--levels", "9", "--levels", "10"}, []int{8, 9, 10}},
		{&C{}, []string{}, nil},
		{&C{[]int{}}, []string{}, []int{}},
		{&C{[]int{11, 12}}, []string{}, []int{11, 12}},
		{&C{[]int{11, 12}}, []string{"--levels", "5"}, []int{5}},
	}
	for _, c := range cases {
		args, err := ParseArgs(c.cfg, c.args)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(args))
		assert.Equal(t, c.expected, c.cfg.Levels)
	}
}

func TestFlagMakerFloatSlice(t *testing.T) {
	type C struct {
		Levels []float64
	}
	cases := []struct {
		cfg      *C
		args     []string
		expected []float64
	}{
		{&C{}, []string{"--levels", "8.9", "--levels", "9.9", "--levels", "10.9"}, []float64{8.9, 9.9, 10.9}},
		{&C{[]float64{}}, []string{"--levels", "8.9", "--levels", "9.9", "--levels", "10.9"}, []float64{8.9, 9.9, 10.9}},
		{&C{}, []string{}, nil},
		{&C{[]float64{}}, []string{}, []float64{}},
		{&C{[]float64{11.3, 12.3}}, []string{}, []float64{11.3, 12.3}},
		{&C{[]float64{11.3, 12.3}}, []string{"--levels", "5.1"}, []float64{5.1}},
	}
	for _, c := range cases {
		args, err := ParseArgs(c.cfg, c.args)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(args))
		assert.Equal(t, c.expected, c.cfg.Levels)
	}
}

func TestFlagMakerInvalidSlice(t *testing.T) {
	type C struct {
		Levels  []int
		Weights []float64
	}
	cases := []struct {
		cfg  *C
		args []string
		l    []int
		w    []float64
	}{
		{
			// invalid flag values won't modify the struct
			&C{Levels: []int{2, 3}, Weights: []float64{2.4, 5.6}},
			[]string{"--levels", "7ax", "--levels", "10", "--weights", "u8.2"},
			[]int{2, 3},
			[]float64{2.4, 5.6},
		},
		{
			// however, valid values before invalid flag values WILL clear slice
			&C{Levels: []int{2, 3}, Weights: []float64{2.4, 5.6}},
			[]string{"--weights", "1.1", "--levels", "10", "--weights", "u8.2", "--levels", "abc"},
			[]int{10},
			[]float64{1.1},
		},
	}

	for _, c := range cases {
		_, err := ParseArgs(c.cfg, c.args)
		assert.Error(t, err)
		assert.Equal(t, c.l, c.cfg.Levels)
		assert.Equal(t, c.w, c.cfg.Weights)
	}
}

func TestFlagMakerVarGet(t *testing.T) {
	var i8 int8 = 3
	var i16 int16 = 4
	var i32 int32 = 5
	var f32 float32 = 10.9
	var u8 uint8 = 22
	var u16 uint16 = 30
	var u32 uint32 = 55
	is := []int{1, 40, 30}
	ss := []string{"haha", "xx"}
	fs := []float64{242.66, 7565.23, 234.67}
	cases := []struct {
		getter   flag.Getter
		expected interface{}
	}{
		{newInt8Value(&i8), i8},
		{newInt16Value(&i16), i16},
		{newInt32Value(&i32), i32},
		{newFloat32Value(&f32), f32},
		{newUint8Value(&u8), u8},
		{newUint16Value(&u16), u16},
		{newUint32Value(&u32), u32},
		{newStringSlice(&ss), ss},
		{newIntSlice(&is), is},
		{newFloat64Slice(&fs), fs},
	}

	for _, c := range cases {
		assert.Equal(t, c.expected, c.getter.Get())
	}
}
