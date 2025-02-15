package event

import (
	"bufio"
	"io"
	"os"
	"strings"
	"unicode"
)

const (
	sshdConfig    = "/etc/ssh/sshd_config"
	SSHD_DATATYPE = 3002
)

var _ Event = (*Sshd)(nil)

type Sshd struct {
	BasicEvent
}

func (Sshd) DataType() int {
	return SSHD_DATATYPE
}

func (Sshd) String() string {
	return "sshdconfig"
}

func (Sshd) Run() (result map[string]interface{}, err error) {
	result = make(map[string]interface{}, 4)
	var file *os.File
	var scan *bufio.Scanner
	if file, err = os.Open(sshdConfig); err != nil {
		return
	}
	defer file.Close()
	// Default value of the configuration
	result["pubkey_authentication"] = "yes"
	result["passwd_authentication"] = "no"
	result["permit_emptypassword"] = "no"
	result["max_authtries"] = "-1" // In my vm, it is 6

	scan = bufio.NewScanner(io.LimitReader(file, 1024*1024))
	for scan.Scan() {
		text := strings.TrimSpace(scan.Text())
		// skip
		if len(text) == 0 || text[:1] == "#" {
			continue
		}
		// Also, according to https://www.cyberciti.biz/faq/create-ssh-config-file-on-linux-unix/
		// "=" is also supported, which is ignored in Elkeid. Just a tidy problem, which can be used
		// in avoiding detection of ssh_config.
		// get PasswordAuthentication & PubkeyAuthentication Only
		fields := strings.FieldsFunc(text, func(c rune) bool {
			return unicode.IsSpace(c) || c == '='
		})
		if len(fields) == 2 {
			switch strings.TrimSpace(fields[0]) {
			case "PasswordAuthentication":
				result["passwd_authentication"] = "passwd_authentication" + "=" + strings.TrimSpace(fields[1])
			case "PubkeyAuthentication":
				result["pubkey_authentication"] = "pubkey_authentication" + "=" + strings.TrimSpace(fields[1])
			case "PermitEmptyPasswords":
				result["permit_emptypassword"] = "permit_emptypassword" + "=" + strings.TrimSpace(fields[1])
			case "MaxAuthTries":
				result["max_auth_tries"] = "max_auth_tries" + "=" + strings.TrimSpace(fields[1])
			}
		}
	}
	return
}

func init() {
	RegistEvent(&Sshd{})
}
