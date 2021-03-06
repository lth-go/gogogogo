package compiler

func fixStatementList(currentBlock *Block, statementList []Statement, fd *FunctionDefinition) {
	for _, statement := range statementList {
		statement.fix(currentBlock, fd)
	}
}

func fixClassMemberExpression(expr *MemberExpression, memberName string) Expression {
	obj := expr.expression

	obj.typeS().fix()

	cd := obj.typeS().classRef.classDefinition

	member := cd.searchMember(memberName)
	if member == nil {
		compileError(expr.Position(), MEMBER_NOT_FOUND_ERR, cd.name, memberName)
	}

	expr.memberDeclaration = member

	switch m := member.(type) {
	case *MethodMember:
		expr.setType(createFunctionDeriveType(m.functionDefinition))
	case *FieldMember:
		expr.setType(m.typeSpecifier)
	}

	return expr

}

// 仅限函数
func fixModuleMemberExpression(expr *MemberExpression, memberName string) Expression {
	innerExpr := expr.expression

	innerExpr.typeS().fix()

	module := innerExpr.(*IdentifierExpression).inner.(*Module)

	moduleCompiler := module.compiler

	fd := moduleCompiler.searchFunction(memberName)
	if fd == nil {
		panic("TODO")
	}

	// TODO 得用当前compiler来添加
	currentCompiler := getCurrentCompiler()

	newExpr := &IdentifierExpression{
		name: memberName,
		inner: &FunctionIdentifier{
			functionDefinition: fd,
			functionIndex:      currentCompiler.addToVmFunctionList(fd),
		},
	}

	newExpr.setType(createFunctionDeriveType(fd))
	newExpr.typeS().fix()

	return newExpr
}
