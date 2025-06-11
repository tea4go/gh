module testlog

go 1.22

replace (
	// WinPC
	github.com/tea4go/gh => C:\MyWork\gitcode\gh
	github.com/tea4go/application/common => C:\MyWork\gitcode\application\common

// WhitePC
//github.com/tea4go/gh => /home/share/mycode/gitcode/gh
//github.com/tea4go/application/common => /home/share/mycode/gitcode/application/common

// Other
//github.com/tea4go/gh =>                 /mnt/share/mycode/gitcode/gh
//github.com/tea4go/application/common => /home/share/mycode/gitcode/application/common

// MacM4
//github.com/tea4go/gh =>                 /Users/tony/mycode/GitCode/gh
//github.com/tea4go/application/common => /Users/tony/mycode/GitCode/application/common
)

require github.com/tea4go/gh v0.0.0-00010101000000-000000000000

require (
	github.com/shiena/ansicolor v0.0.0-20230509054315-a9deabde6e02 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/net v0.22.0 // indirect
)
