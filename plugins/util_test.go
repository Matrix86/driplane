package plugins

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestUtilPluginSleepMethod(t *testing.T) {
	u := GetUtil()

	res := u.Sleep(0)
	if res.Status == false {
		t.Errorf("bad response: expected=%t had=%t", true, res.Status)
	}
}

func TestUtilPluginGetEnvMethod(t *testing.T) {
	u := GetUtil()

	res := u.Getenv("ENVTESTVAR")
	if res.Value != "" {
		t.Errorf("the env var should be empty")
	}

	os.Setenv("ENVTESTVAR", "1")

	res = u.Getenv("ENVTESTVAR")
	if res.Value != "1" {
		t.Errorf("the env var should contain '1'")
	}
}

func TestUtilPluginMd5Method(t *testing.T) {
	u := GetUtil()

	file, err := os.CreateTemp(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file")
	}
	defer os.Remove(file.Name())

	res := u.Md5File("/tmp/notexistentfile_")
	if res.Status == true {
		t.Errorf("Status should be false")
	}

	res = u.Md5File(file.Name())
	if res.Status == false {
		t.Errorf("Status should be true")
	}

	expected := "d41d8cd98f00b204e9800998ecf8427e"
	if res.Value != expected {
		t.Errorf("Value has a bad value: expected=%s had=%s", expected, res.Value)
	}
}

func TestUtilPluginSha1Method(t *testing.T) {
	u := GetUtil()

	file, err := os.CreateTemp(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file")
	}
	defer os.Remove(file.Name())

	res := u.Sha1File("/tmp/notexistentfile_")
	if res.Status == true {
		t.Errorf("Status should be false")
	}

	res = u.Sha1File(file.Name())
	if res.Status == false {
		t.Errorf("Status should be true")
	}

	expected := "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	if res.Value != expected {
		t.Errorf("Value has a bad value: expected=%s had=%s", expected, res.Value)
	}
}

func TestUtilPluginSha256Method(t *testing.T) {
	u := GetUtil()

	file, err := os.CreateTemp(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file")
	}
	defer os.Remove(file.Name())

	res := u.Sha256File("/tmp/notexistentfile_")
	if res.Status == true {
		t.Errorf("Status should be false")
	}

	res = u.Sha256File(file.Name())
	if res.Status == false {
		t.Errorf("Status should be true")
	}

	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if res.Value != expected {
		t.Errorf("Value has a bad value: expected=%s had=%s", expected, res.Value)
	}
}

func TestUtilPluginSha512Method(t *testing.T) {
	u := GetUtil()

	file, err := os.CreateTemp(os.TempDir(), "prefix")
	if err != nil {
		t.Errorf("cannot create a temporary file")
	}
	defer os.Remove(file.Name())

	res := u.Sha512File("/tmp/notexistentfile_")
	if res.Status == true {
		t.Errorf("Status should be false")
	}

	res = u.Sha512File(file.Name())
	if res.Status == false {
		t.Errorf("Status should be true")
	}

	expected := "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
	if res.Value != expected {
		t.Errorf("Value has a bad value: expected=%s had=%s", expected, res.Value)
	}
}

func TestUtilPluginExecCommandMethod(t *testing.T) {
	u := GetUtil()

	var echoCmd []string
	expectedOutput := "hello world"
	if runtime.GOOS == "windows" {
		echoCmd = []string{"cmd", "/C", "echo", expectedOutput}
	} else {
		echoCmd = []string{"echo", expectedOutput}
	}

	tests := []struct {
		name           string   // test name
		cmdParts       []string // command
		inputData      string   // stdin
		expectedStatus bool     // status field of the response
		expectError    bool     // Error field of the response
		expectedValue  string   // returned value of the command
	}{
		{
			name:           "Error: no command ",
			cmdParts:       []string{},
			inputData:      "",
			expectedStatus: false,
			expectError:    true,
			expectedValue:  "",
		},
		{
			name:           "Success: Simple Echo command",
			cmdParts:       echoCmd,
			inputData:      "",
			expectedStatus: true,
			expectError:    false,
			expectedValue:  expectedOutput,
		},
		{
			name:           "Error: Command not exist",
			cmdParts:       []string{"not_existing_command_12345"},
			inputData:      "",
			expectedStatus: false,
			expectError:    true,
			expectedValue:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := u.ExecCommand(tt.cmdParts, tt.inputData)

			if response.Status != tt.expectedStatus {
				t.Errorf("Wrong status: expected %v, obtained %v", tt.expectedStatus, response.Status)
			}

			if tt.expectError && response.Error == nil {
				t.Errorf("Expected error but obtained nil")
			}
			if !tt.expectError && response.Error != nil {
				t.Errorf("Error not expected but obtained: %v", response.Error)
			}

			cleanOutput := strings.TrimSpace(response.Value)

			if tt.expectedValue != "" {

				if !strings.Contains(cleanOutput, tt.expectedValue) {
					t.Errorf("Wrong output. \n Expected: %q\n Obtained: %q", tt.expectedValue, cleanOutput)
				}
			}
		})
	}

}
