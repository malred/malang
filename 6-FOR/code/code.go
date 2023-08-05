// code/code.go
package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// 操作码定义
const (
	OpConstant Opcode = iota
	OpAdd             // +
	OpSub             // -
	OpMul             // *
	OpDiv             // /
	OpPop
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpMinus         // - 负号
	OpBang          // !
	OpJumpNotTruthy // 有条件跳转
	OpJump          // 无条件跳转
	OpNull          // 将Null压栈
	OpGetGlobal     // 从全局存储中取值
	OpSetGlobal     // 向全局存储中存值
	OpArray         // 构建数组
	OpHash          // 构建哈希
	OpIndex         // 索引运算
)

type Instructions []byte

type Opcode byte

// 操作码定义信息
type Definition struct {
	Name          string
	OperandWidths []int
}

// 操作码定义详细信息
var definitions = map[Opcode]*Definition{
	OpConstant:    {"OpConstant", []int{2}},
	OpAdd:         {"OpAdd", []int{}},
	OpSub:         {"OpSub", []int{}},
	OpMul:         {"OpMul", []int{}},
	OpDiv:         {"OpDiv", []int{}},
	OpPop:         {"OpPop", []int{}},
	OpTrue:        {"OpTrue", []int{}},
	OpFalse:       {"OpFalse", []int{}},
	OpEqual:       {"OpEqual", []int{}},
	OpNotEqual:    {"OpNotEqual", []int{}},
	OpGreaterThan: {"OpGreaterThan", []int{}},
	OpMinus:       {"OpMinus", []int{}},
	OpBang:        {"OpBang", []int{}},
	// 跳转指令有两字节大小（16位）的操作数（目标指令的绝对偏移量）
	OpJumpNotTruthy: {"OpJumpNotTruthy", []int{2}},
	OpJump:          {"OpJump", []int{2}},
	OpNull:          {"OpNull", []int{}},
	OpGetGlobal:     {"OpGetGlobal", []int{2}},
	OpSetGlobal:     {"OpSetGlobal", []int{2}},
	OpArray:         {"OpArray", []int{2}},
	OpHash:          {"OpHash", []int{2}},
	OpIndex:         {"OpIndex", []int{}},
}

// 查看操作码定义
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// 快速构建单字节码指令
func Make(op Opcode, operands ...int) []byte {
	// 从已定义的操作码中寻找
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	// 根据操作数的个数决定需要返回的[]byte长度（操作码+操作数）
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	// 返回单字节码编译后的机器指令
	instruction := make([]byte, instructionLen)
	// 操作码占首1字节
	instruction[0] = byte(op)

	// 第一位是操作码，已经放入
	offset := 1
	// 遍历操作数
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			// 操作数大端编码为uint16到instruction
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		}
		// 向后移动，继续遍历操作数
		offset += width
	}

	return instruction
}

// 更好地打印字节码指令
func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		// 读取操作数，read是读取了多少字节
		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		// 向后移动read个字节
		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n", len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}

// 逆make
// 反编码make编码后的操作数
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		}

		offset += width
	}

	return operands, offset
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}
