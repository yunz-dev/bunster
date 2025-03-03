package ir

import (
	"fmt"
)

type Compare struct {
	Left     Instruction
	Operator string
	Right    Instruction
}

func (c Compare) togo() string {
	return fmt.Sprintf(
		`if %s %s %s {
			shell.ExitCode = 0 
		} else {
			shell.ExitCode = 1
		}
		`, c.Left.togo(), c.Operator, c.Right.togo())
}

type CompareArithmetics struct {
	Left     Instruction
	Operator string
	Right    Instruction
}

func (c CompareArithmetics) togo() string {
	return fmt.Sprintf(
		`if runtime.NumberCompare(%s, %q, %s) {
			shell.ExitCode = 0 
		} else {
			shell.ExitCode = 1
		}
		`, c.Left.togo(), c.Operator, c.Right.togo())
}

type TestFilesHaveSameDevAndInoNumbers struct {
	File1 Instruction
	File2 Instruction
}

func (c TestFilesHaveSameDevAndInoNumbers) togo() string {
	return fmt.Sprintf(
		`if runtime.FilesHaveSameDevAndIno(%s, %s) {
			shell.ExitCode = 0 
		} else {
			shell.ExitCode = 1
		}
		`, c.File1.togo(), c.File2.togo())
}

type FileIsOlderThan struct {
	File1 Instruction
	File2 Instruction
}

func (c FileIsOlderThan) togo() string {
	return fmt.Sprintf(
		`if runtime.FileIsOlderThan(%s, %s) {
			shell.ExitCode = 0 
		} else {
			shell.ExitCode = 1
		}
		`, c.File1.togo(), c.File2.togo())
}

type TestStringIsIsNotZero struct {
	String Instruction
}

func (c TestStringIsIsNotZero) togo() string {
	return fmt.Sprintf(
		`if len(%s) == 0 {
			shell.ExitCode = 1 
		} else {
			shell.ExitCode = 0
		}
		`, c.String.togo())
}
