package broccli

import (
	"os"
	"testing"
)

// TestParamValidationBasic tests basic validation for specific types.
func TestParamValidationBasic(t *testing.T) {
	t.Parallel()

	p := &param{}
	if p.validateValue("") != nil {
		t.Errorf("Empty param should validate")
	}

	for _, typ := range []int64{TypeInt, TypeFloat, TypeAlphanumeric, TypePathFile} {
		p.valueType = typ
		if p.validateValue("") != nil {
			t.Errorf("Empty param should validate")
		}
	}

	p.valueType = TypeInt
	if p.validateValue("48") != nil {
		t.Errorf("Int param should validate")
	}

	if p.validateValue("aa") == nil {
		t.Errorf("Int param should not validate string")
	}

	p.valueType = TypeFloat
	if p.validateValue("48.998") != nil {
		t.Errorf("Float param should validate")
	}

	if p.validateValue("48") == nil {
		t.Errorf("Float param should not validate int")
	}

	if p.validateValue("aa") == nil {
		t.Errorf("Float param should not validate string")
	}

	p.valueType = TypeAlphanumeric
	if p.validateValue("a123aaAEz") != nil {
		t.Errorf("Alphanumeric param should validate")
	}

	if p.validateValue("a.z") == nil {
		t.Errorf("Alphanumeric param should not validate")
	}

	p.valueType = TypePathFile
	if p.validateValue("anything/here") != nil {
		t.Errorf("TypePathFile param should validate")
	}
}

// TestParamValidationRequired tests IsRequired flag.
func TestParamValidationRequired(t *testing.T) {
	t.Parallel()

	p := &param{
		flags: IsRequired,
	}
	if p.validateValue("") == nil {
		t.Errorf("Empty param with IsRequired should not validate")
	}

	if p.validateValue("aa") != nil {
		t.Errorf("Param with IsRequired should validate")
	}
}

// TestParamValidationExtraChars tests flags such as AllowUnderscore that allow TypeAlphanumeric to contain
// extra characters.
func TestParamValidationExtraChars(t *testing.T) {
	t.Parallel()

	p := &param{
		valueType: TypeAlphanumeric,
		flags:     AllowDots,
	}
	if p.validateValue("aZ09.az") != nil {
		t.Errorf("Alphanumeric param with extra chars should validate")
	}

	p.flags = AllowUnderscore
	if p.validateValue("aZ09_09") != nil {
		t.Errorf("Alphanumeric param with extra chars should validate")
	}

	p.flags = AllowHyphen
	if p.validateValue("aZ09-09") != nil {
		t.Errorf("Alphanumeric param with extra chars should validate")
	}

	if p.validateValue("aZ09-_09") == nil {
		t.Errorf("Alphanumeric param with extra chars should fail")
	}

	p.flags = AllowUnderscore | AllowHyphen
	if p.validateValue("aZ09_0-9") != nil {
		t.Errorf("Alphanumeric param with extra chars should validate")
	}
}

// TestParamValidationMultipleValues tests params that allow multiple values.
func TestParamValidationMultipleValues(t *testing.T) {
	t.Parallel()

	p := &param{
		valueType: TypeAlphanumeric,
		flags:     AllowMultipleValues | AllowDots,
	}
	if p.validateValue("aZ09.az,x00,y2.3") != nil {
		t.Errorf("Alphanumeric param with extra chars and multiple values should validate")
	}

	if p.validateValue("aZ09.az,x-00,y2.3") == nil {
		t.Errorf("Alphanumeric param with extra chars and multiple values should fail")
	}

	p.flags = AllowMultipleValues | AllowHyphen | AllowUnderscore | SeparatorSemiColon
	if p.validateValue("aZ09_-az;x0-0;y2_-3") != nil {
		t.Errorf("Alphanumeric param with extra chars and multiple values should validate")
	}

	if p.validateValue("aZ09az,x-00,y2_3") == nil {
		t.Errorf("Alphanumeric param with extra chars and multiple values should fail")
	}

	p.valueType = TypeFloat
	if p.validateValue("5.8;2.3") != nil {
		t.Errorf("Alphanumeric param with extra chars and multiple values should validate")
	}
}

// TestParamValidationFiles creates param of TypePathFile and tests additional validation flags related to checking
// if file is a regular file, if it exists etc.
func TestParamValidationFiles(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "example")
	if err != nil {
		t.Errorf("error creating temporary file")
	}

	defer func() {
		err := os.Remove(f.Name())
		if err != nil {
			t.Errorf("error removing temporary file")
		}
	}()

	p := &param{
		valueType: TypePathFile,
	}
	if p.validateValue("/non-existing/path") != nil {
		t.Errorf("PathFile param should validate")
	}

	p.flags = IsNotExistent
	if p.validateValue("/non-existing/path") != nil {
		t.Errorf("PathFile param with IsNotExistent should validate")
	}

	if p.validateValue(f.Name()) == nil {
		t.Errorf("PathFile param with IsNotExistent should fail")
	}

	if p.validateValue("") != nil {
		t.Errorf("Empty PathFile param with IsNotExistent should validate")
	}

	p.flags = IsExistent
	if p.validateValue("/non-existing/path") == nil {
		t.Errorf("PathFile param with IsExistent should fail")
	}

	if p.validateValue(f.Name()) != nil {
		t.Errorf("PathFile param with IsNotExistent should validate")
	}

	if p.validateValue("") != nil {
		t.Errorf("Empty PathFile param with IsExistent should validate")
	}

	p.flags = IsRegularFile
	if p.validateValue("") != nil {
		t.Errorf("Empty PathFile param with IsRegularFile should validate")
	}

	if p.validateValue(f.Name()) != nil {
		t.Errorf("PathFile param with IsRegularFile should validate")
	}

	p.flags = IsDirectory
	if p.validateValue("") != nil {
		t.Errorf("Empty PathFile param with IsDirectory should validate")
	}

	if p.validateValue("") != nil {
		t.Errorf("Empty PathFile param with IsDirectory should validate")
	}

	p.flags = IsExistent | IsValidJSON
	if p.validateValue(f.Name()) != nil {
		t.Errorf("PathFile param with IsExistent and IsValidJSON should validate")
	}

	p.flags = IsExistent | IsRegularFile | IsValidJSON
	if p.validateValue(f.Name()) == nil {
		t.Errorf("PathFile param with IsExistent should fail")
	}

	_, err = f.WriteString("{\"valid\":\"json\"}")
	if err != nil {
		t.Errorf("error writing to file")
	}

	p.flags = IsExistent | IsRegularFile | IsValidJSON
	if p.validateValue(f.Name()) != nil {
		t.Errorf("PathFile param with IsExistent should validate")
	}

	_, err = f.WriteString("{\"valid\":\"json\"}")
	if err != nil {
		t.Errorf("error writing to file")
	}

	err = f.Close()
	if err != nil {
		t.Errorf("error closing file")
	}

	p.flags = IsExistent | IsRegularFile | IsValidJSON
	if p.validateValue(f.Name()) == nil {
		t.Errorf("PathFile param with IsExistent should fail")
	}
}
