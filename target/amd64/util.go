package amd64

func isPow2(x int) bool {
	return ((x != 0) && ((x & (^x + 1)) == x))
}

func isRegSizeBits(bits int) bool {
	return bits == 8 || bits == 16 || bits == 32 || bits == 64
}

func sizeSuffixBits(x int) string {
	switch x {
	case 8:
		return "b"
	case 16:
		return "w"
	case 32:
		return "l"
	case 64:
		return "q"
	default:
		panic("internal error: tried to get suffix for invalid size")
	}
}

var registers = [][]string{
	{"rax", "eax", "ax", "al"},
	{"rbx", "ebx", "bx", "bl"},
	{"rcx", "ecx", "cx", "cl"},
	{"rdx", "edx", "dx", "dl"},
	{"rdi", "edi", "di", "dil"},
	{"rsi", "esi", "si", "sil"},
	{"r8", "r8d", "r8w", "r8b"},
	{"r9", "r9d", "r9w", "r9b"},
	{"r10", "r10d", "r10w", "r10b"},
	{"r11", "r11d", "r11w", "r11b"},
	{"r12", "r12d", "r12w", "r12b"},
	{"r13", "r13d", "r13w", "r13b"},
	{"r14", "r14d", "r14w", "r14b"},
	{"r15", "r15d", "r15w", "r15b"},
}

func regToSize(ireg string, bits int) string {
	for _, regList := range registers {
		for _, reg := range regList {
			if reg == ireg {
				switch bits {
				case 64:
					return regList[0]
				case 32:
					return regList[1]
				case 16:
					return regList[2]
				case 8:
					return regList[3]
				default:
					panic("")
				}
			}
		}
	}

	panic("")
}
