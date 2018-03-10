package compiler

func fixStatementList(currentBlock *Block, statementList []Statement, fd *FunctionDefinition) {
	for _, statement := range statementList {
		statement.fix(currentBlock, fd)
	}
}

func fixClassMemberExpression(expr *MemberExpression, obj Expression, memberName string) Expression {

	obj.typeS().fix()

	cd := obj.typeS().classRef.classDefinition

	member := cd.searchMember(memberName)
	if member == nil {
		compileError(expr.Position(), MEMBER_NOT_FOUND_ERR, cd.name, memberName)
	}

	expr.declaration = member

	switch m := member.(type) {
	case *MethodMember:
		expr.setType(createFunctionDeriveType(m.functionDefinition))
	case *FieldMember:
		expr.setType(m.typeSpecifier)
	}

	return expr

}
