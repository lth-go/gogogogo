package compiler

import (
	"github.com/lth-go/gogogogo/vm"
)

// ==============================
// Statement 接口
// ==============================

// Statement 语句接口
type Statement interface {
	// Pos接口
	Pos

	fix(*Block, *FunctionDefinition)
	generate(exe *vm.Executable, currentBlock *Block, ob *OpCodeBuf)

	show(indent int)
}

type StatementImpl struct {
	PosImpl
}

// ==============================
// ExpressionStatement
// ==============================

// ExpressionStatement 表达式语句
type ExpressionStatement struct {
	StatementImpl
	expression Expression
}

func (stmt *ExpressionStatement) show(indent int) {
	printWithIndent("ExprStmt", indent)

	subIndent := indent + 2

	stmt.expression.show(subIndent)
}

func (stmt *ExpressionStatement) fix(currentBlock *Block, fd *FunctionDefinition) {
	stmt.expression = stmt.expression.fix(currentBlock)
}

func (stmt *ExpressionStatement) generate(exe *vm.Executable, currentBlock *Block, ob *OpCodeBuf) {
	expr := stmt.expression
	switch assignExpr := expr.(type) {
	case *AssignExpression:
		// TODO
		assignExpr.generateEx(exe, currentBlock, ob, true)
	default:
		expr.generate(exe, currentBlock, ob)
		ob.generateCode(expr.Position(), vm.VM_POP)
	}
}

// ==============================
// IfStatement
// ==============================

//
// Elif
//
type Elif struct {
	condition Expression
	block     *Block
}

// IfStatement if表达式
type IfStatement struct {
	StatementImpl

	condition Expression
	thenBlock *Block
	elifList  []*Elif
	elseBlock *Block
}

func (stmt *IfStatement) show(indent int) {
	printWithIndent("IfStmt", indent)

	subIndent := indent + 2
	stmt.condition.show(subIndent)
	if stmt.thenBlock != nil {
		stmt.thenBlock.show(subIndent)
	}
	for _, elif := range stmt.elifList {
		printWithIndent("Elif", subIndent)
		elif.condition.show(subIndent + 2)
		elif.block.show(subIndent + 2)
	}

	if stmt.elseBlock != nil {
		stmt.elseBlock.show(subIndent)
	}
}

func (stmt *IfStatement) fix(currentBlock *Block, fd *FunctionDefinition) {

	stmt.condition = stmt.condition.fix(currentBlock)

	if !isBoolean(stmt.condition.typeS()) {
		compileError(stmt.condition.Position(), IF_CONDITION_NOT_BOOLEAN_ERR)
	}

	if stmt.thenBlock != nil {
		fixStatementList(stmt.thenBlock, stmt.thenBlock.statementList, fd)
	}

	for _, elif := range stmt.elifList {
		elif.condition = elif.condition.fix(currentBlock)

		if elif.block != nil {
			fixStatementList(elif.block, elif.block.statementList, fd)
		}
	}

	if stmt.elseBlock != nil {
		fixStatementList(stmt.elseBlock, stmt.elseBlock.statementList, fd)
	}
}

func (stmt *IfStatement) generate(exe *vm.Executable, currentBlock *Block, ob *OpCodeBuf) {

	stmt.condition.generate(exe, currentBlock, ob)

	// 获取false跳转地址
	ifFalseLabel := ob.getLabel()
	ob.generateCode(stmt.Position(), vm.VM_JUMP_IF_FALSE, ifFalseLabel)

	if stmt.thenBlock != nil {
		generateStatementList(exe, stmt.thenBlock, stmt.thenBlock.statementList, ob)
	}

	// 获取结束跳转地址
	endLabel := ob.getLabel()

	// 直接跳到最后
	ob.generateCode(stmt.Position(), vm.VM_JUMP, endLabel)

	// 设置false跳转地址,如果false,直接执行这里
	ob.setLabel(ifFalseLabel)

	for _, elif := range stmt.elifList {
		elif.condition.generate(exe, currentBlock, ob)

		// 获取false跳转地址
		ifFalseLabel = ob.getLabel()
		ob.generateCode(stmt.Position(), vm.VM_JUMP_IF_FALSE, ifFalseLabel)

		generateStatementList(exe, elif.block, elif.block.statementList, ob)

		// 直接跳到最后
		ob.generateCode(stmt.Position(), vm.VM_JUMP, endLabel)

		// 设置false跳转地址,如果false,直接执行这里
		ob.setLabel(ifFalseLabel)
	}

	if stmt.elseBlock != nil {
		generateStatementList(exe, stmt.elseBlock, stmt.elseBlock.statementList, ob)
	}

	// 设置结束地址
	ob.setLabel(endLabel)
}

// ==============================
// ForStatement
// ==============================

// ForStatement for语句
type ForStatement struct {
	StatementImpl

	init      Expression
	condition Expression
	post      Expression
	block     *Block
}

func (stmt *ForStatement) show(indent int) {
	printWithIndent("ForStmt", indent)
	subIndent := indent + 2

	if stmt.init != nil {
		stmt.init.show(subIndent)
	}
	if stmt.condition != nil {
		stmt.condition.show(subIndent)
	}
	if stmt.post != nil {
		stmt.post.show(subIndent)
	}

	if stmt.block != nil {
		stmt.block.show(subIndent)
	}
}

func (stmt *ForStatement) fix(currentBlock *Block, fd *FunctionDefinition) {
	if stmt.init != nil {
		stmt.init = stmt.init.fix(currentBlock)
	}

	if stmt.condition != nil {
		stmt.condition = stmt.condition.fix(currentBlock)

		if !isBoolean(stmt.condition.typeS()) {
			compileError(stmt.condition.Position(), FOR_CONDITION_NOT_BOOLEAN_ERR)
		}
	}

	if stmt.post != nil {
		stmt.post = stmt.post.fix(currentBlock)
	}

	if stmt.block != nil {
		fixStatementList(stmt.block, stmt.block.statementList, fd)
	}
}
func (stmt *ForStatement) generate(exe *vm.Executable, currentBlock *Block, ob *OpCodeBuf) {

	if stmt.init != nil {
		stmt.init.generate(exe, currentBlock, ob)
	}

	// 获取循环地址
	loopLabel := ob.getLabel()

	// 设置循环地址
	ob.setLabel(loopLabel)

	if stmt.condition != nil {
		stmt.condition.generate(exe, currentBlock, ob)
	}

	label := ob.getLabel()

	if stmt.condition != nil {
		// 如果条件为否,跳转到break, label = parent.breakLabel
		ob.generateCode(stmt.Position(), vm.VM_JUMP_IF_FALSE, label)
	}

	if stmt.block != nil {
		parent := stmt.block.parent.(*StatementBlockInfo)
		// 获取break,continue地址
		parent.breakLabel = label
		parent.continueLabel = label

		generateStatementList(exe, stmt.block, stmt.block.statementList, ob)
	}

	// 如果有continue,直接跳过block,从这里执行, label = parent.continueLabel
	ob.setLabel(label)

	if stmt.post != nil {
		stmt.post.generate(exe, currentBlock, ob)
	}

	// 跳回到循环开头
	ob.generateCode(stmt.Position(), vm.VM_JUMP, loopLabel)

	// 设置结束标签, label = parent.breakLabel
	ob.setLabel(label)
}

// ==============================
// ReturnStatement
// ==============================

// ReturnStatement return 语句
type ReturnStatement struct {
	StatementImpl

	// 返回值
	returnValue Expression
}

func (stmt *ReturnStatement) show(indent int) {
	printWithIndent("ReturnStmt", indent)
	subIndent := indent + 2

	stmt.returnValue.show(subIndent)
}

func (stmt *ReturnStatement) fix(currentBlock *Block, fd *FunctionDefinition) {

	fdType := fd.typeS()

	// 如果没有返回值,添加之
	if stmt.returnValue != nil {
		if fdType.deriveList == nil && isVoid(fdType) {
			compileError(stmt.Position(), RETURN_IN_VOID_FUNCTION_ERR)
		}

		stmt.returnValue = stmt.returnValue.fix(currentBlock)

		// 类型转换
		stmt.returnValue = createAssignCast(stmt.returnValue, fdType)

		return
	}

	// return value == nil

	// 衍生类型
	if fdType.deriveList != nil {
		if !fdType.isArrayDerive() {
			panic("TODO")
		}
		stmt.returnValue = createNullExpression(stmt.Position())

		return
	}

	// 基础类型
	switch fdType.basicType {
	case vm.VoidType:
		stmt.returnValue = createIntExpression(stmt.Position())
	case vm.BooleanType:
		stmt.returnValue = createBooleanExpression(stmt.Position())
	case vm.IntType:
		stmt.returnValue = createIntExpression(stmt.Position())
	case vm.DoubleType:
		stmt.returnValue = createDoubleExpression(stmt.Position())
	case vm.StringType:
		stmt.returnValue = createStringExpression(stmt.Position())
	case vm.ClassType:
		stmt.returnValue = createNullExpression(stmt.Position())
	case vm.NullType:
		fallthrough
	default:
		panic("TODO")
	}
}

func (stmt *ReturnStatement) generate(exe *vm.Executable, currentBlock *Block, ob *OpCodeBuf) {
	if stmt.returnValue == nil {
		panic("Return value is nil.")
	}

	stmt.returnValue.generate(exe, currentBlock, ob)

	ob.generateCode(stmt.Position(), vm.VM_RETURN)
}

// ==============================
// BreakStatement
// ==============================

// BreakStatement break 语句
type BreakStatement struct {
	StatementImpl
}

func (stmt *BreakStatement) show(indent int) {
	printWithIndent("BreakStmt", indent)
}

func (stmt *BreakStatement) fix(currentBlock *Block, fd *FunctionDefinition) {}

func (stmt *BreakStatement) generate(exe *vm.Executable, currentBlock *Block, ob *OpCodeBuf) {
	// 向外寻找,直到找到for的block
	for block := currentBlock; block != nil; block = block.outerBlock {
		switch block.parent.(type) {
		case *StatementBlockInfo:
			ob.generateCode(stmt.Position(), vm.VM_JUMP, block.parent.(*StatementBlockInfo).breakLabel)
			return
		default:
			continue
		}
	}
	compileError(stmt.Position(), LABEL_NOT_FOUND_ERR)
}

// ==============================
// ContinueStatement
// ==============================

// ContinueStatement continue 语句
type ContinueStatement struct {
	StatementImpl
}

func (stmt *ContinueStatement) show(indent int) {
	printWithIndent("ContinueStmt", indent)
}

func (stmt *ContinueStatement) fix(currentBlock *Block, fd *FunctionDefinition) {}

func (stmt *ContinueStatement) generate(exe *vm.Executable, currentBlock *Block, ob *OpCodeBuf) {
	// 向外寻找,直到找到for的block
	for block := currentBlock; block != nil; block = block.outerBlock {
		switch block.parent.(type) {
		case *StatementBlockInfo:
			ob.generateCode(stmt.Position(), vm.VM_JUMP, block.parent.(*StatementBlockInfo).continueLabel)
			return
		default:
			continue
		}
	}
	compileError(stmt.Position(), LABEL_NOT_FOUND_ERR)

}

// ==============================
// Declaration
// ==============================

// Declaration 声明语句
type Declaration struct {
	StatementImpl

	typeSpecifier *TypeSpecifier

	name        string
	initializer Expression

	variableIndex int

	isLocal bool
}

func (stmt *Declaration) show(indent int) {
	printWithIndent("DeclStmt", indent)

	subIndent := indent + 2
	if stmt.initializer != nil {
		stmt.initializer.show(subIndent)
	}
}

func (stmt *Declaration) fix(currentBlock *Block, fd *FunctionDefinition) {
	currentBlock.addDeclaration(stmt, fd, stmt.Position())

	stmt.typeSpecifier.fix()

	// 类型转换
	if stmt.initializer != nil {
		stmt.initializer = stmt.initializer.fix(currentBlock)
		stmt.initializer = createAssignCast(stmt.initializer, stmt.typeSpecifier)
	}
}

func (stmt *Declaration) generate(exe *vm.Executable, currentBlock *Block, ob *OpCodeBuf) {
	if stmt.initializer == nil {
		return
	}

	stmt.initializer.generate(exe, currentBlock, ob)
	generatePopToIdentifier(stmt, stmt.Position(), ob)
}
