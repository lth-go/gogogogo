%{
package compiler

import (
    "github.com/lth-go/gogogogo/vm"
    "strconv"
)
%}

%union{
    parameter            *Parameter
    parameter_list       []*Parameter

    statement            Statement
    statement_list       []Statement

    expression           Expression
    expression_list      []Expression

    block                *Block
    else_if              []*ElseIf

    type_specifier       *TypeSpecifier

    import_spec          *ImportSpec
    import_spec_list     []*ImportSpec

    function_definition  *FunctionDefinition

    tok                  Token
}

%token<tok> IF ELSE FOR RETURN_T BREAK CONTINUE
        LP RP LC RC LB RB
        SEMICOLON COMMA COLON
        ASSIGN_T
        LOGICAL_AND LOGICAL_OR
        EQ NE GT GE LT LE
        ADD SUB MUL DIV
        INT_LITERAL FLOAT_LITERAL STRING_LITERAL TRUE_T FALSE_T
        NULL_T
        IDENTIFIER
        EXCLAMATION DOT
        VOID_T BOOL_T INT_T FLOAT_T STRING_T
        IMPORT
        VAR
        FUNC
        TYPE
        STRUCT

%type <import_spec> import_declaration
%type <import_spec_list> import_list

%type <expression> expression expression_opt
      assignment_expression
      logical_and_expression logical_or_expression
      equality_expression relational_expression
      additive_expression multiplicative_expression
      unary_expression postfix_expression primary_expression primary_no_new_array
      array_literal
%type <expression_list> expression_list argument_list

%type <statement> statement
      if_statement for_statement
      return_statement break_statement continue_statement
      declaration_statement
%type <statement_list> statement_list
/* TODO: 临时处理 */
%type <parameter> receiver_or_nil
%type <parameter_list> parameter_list
%type <block> block
%type <else_if> else_if
%type <type_specifier> basic_type_specifier type_specifier array_type_specifier

%%

translation_unit
        : initial_declaration definition_or_statement
        | translation_unit definition_or_statement
        ;
initial_declaration
        : /* empty */
        {
            setImportList(nil)
        }
        | import_list
        {
            setImportList($1)
        }
        ;
import_list
        : import_declaration
        {
            $$ = createImportSpecList($1)
        }
        | import_list import_declaration
        {
            $$ = append($1, $2)
        }
        ;
import_declaration
        : IMPORT STRING_LITERAL SEMICOLON
        {
            $$ = createImportSpec($2.Lit)
        }
        ;
definition_or_statement
        : function_definition
        | statement
        {
            l := yylex.(*Lexer)
            l.compiler.statementList = append(l.compiler.statementList, $1)
        }
        ;
basic_type_specifier
        : VOID_T
        {
            $$ = createTypeSpecifier(vm.VoidType, $1.Position())
        }
        | BOOL_T
        {
            $$ = createTypeSpecifier(vm.BooleanType, $1.Position())
        }
        | INT_T
        {
            $$ = createTypeSpecifier(vm.IntType, $1.Position())
        }
        | FLOAT_T
        {
            $$ = createTypeSpecifier(vm.DoubleType, $1.Position())
        }
        | STRING_T
        {
            $$ = createTypeSpecifier(vm.StringType, $1.Position())
        }
        ;
/* TODO: LB RB type_specifier */
array_type_specifier
        : LB RB basic_type_specifier
        {
            $$ = createArrayTypeSpecifier($3)
            $$.SetPosition($1.Position())
        }
        | LB RB array_type_specifier
        {
            $$ = createArrayTypeSpecifier($3)
        }
        ;
type_specifier
        : basic_type_specifier
        {
            $$ = $1
        }
        | array_type_specifier
        ;
function_definition
        : FUNC receiver_or_nil IDENTIFIER LP parameter_list RP type_specifier block
        {
            l := yylex.(*Lexer)
            l.compiler.functionDefine($7, $3.Lit, $5, $8)
        }
        | FUNC receiver_or_nil IDENTIFIER LP RP type_specifier block
        {
            l := yylex.(*Lexer)
            l.compiler.functionDefine($6, $3.Lit, []*Parameter{}, $7)
        }
        | FUNC receiver_or_nil IDENTIFIER LP parameter_list RP type_specifier SEMICOLON
        {
            l := yylex.(*Lexer)
            l.compiler.functionDefine($7, $3.Lit, $5, nil)
        }
        | FUNC receiver_or_nil IDENTIFIER LP RP type_specifier SEMICOLON
        {
            l := yylex.(*Lexer)
            l.compiler.functionDefine($6, $3.Lit, []*Parameter{}, nil)
        }
        ;
receiver_or_nil
        :
        {
            $$ = nil
        }
        | LP IDENTIFIER type_specifier RP
        {
            $$ = &Parameter{typeSpecifier: $3, name: $2.Lit}
        }
        ;
parameter_list
        : IDENTIFIER type_specifier
        {
            parameter := &Parameter{typeSpecifier: $2, name: $1.Lit}
            $$ = []*Parameter{parameter}
        }
        | parameter_list COMMA IDENTIFIER type_specifier
        {
            $$ = append($1, &Parameter{typeSpecifier: $4, name: $3.Lit})
        }
        ;
argument_list
        : assignment_expression
        {
            $$ = []Expression{$1}
        }
        | argument_list COMMA assignment_expression
        {
            $$ = append($1, $3)
        }
        ;
statement_list
        : statement
        {
            $$ = []Statement{$1}
        }
        | statement_list statement
        {
            $$ = append($1, $2)
        }
        ;
expression
        : assignment_expression
        | expression COMMA assignment_expression
        {
            $$ = &CommaExpression{left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        ;
assignment_expression
        : logical_or_expression
        | primary_expression ASSIGN_T assignment_expression
        {
            $$ = &AssignExpression{left: $1, operand: $3}
            $$.SetPosition($1.Position())
        }
        ;
logical_or_expression
        : logical_and_expression
        | logical_or_expression LOGICAL_OR logical_and_expression
        {
            $$ = &BinaryExpression{operator: LogicalOrOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        ;
logical_and_expression
        : equality_expression
        | logical_and_expression LOGICAL_AND equality_expression
        {
            $$ = &BinaryExpression{operator: LogicalAndOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        ;
equality_expression
        : relational_expression
        | equality_expression EQ relational_expression
        {
            $$ = &BinaryExpression{operator: EqOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        | equality_expression NE relational_expression
        {
            $$ = &BinaryExpression{operator: NeOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        ;
relational_expression
        : additive_expression
        | relational_expression GT additive_expression
        {
            $$ = &BinaryExpression{operator: GtOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        | relational_expression GE additive_expression
        {
            $$ = &BinaryExpression{operator: GeOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        | relational_expression LT additive_expression
        {
            $$ = &BinaryExpression{operator: LtOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        | relational_expression LE additive_expression
        {
            $$ = &BinaryExpression{operator: LeOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        ;
additive_expression
        : multiplicative_expression
        | additive_expression ADD multiplicative_expression
        {
            $$ = &BinaryExpression{operator: AddOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        | additive_expression SUB multiplicative_expression
        {
            $$ = &BinaryExpression{operator: SubOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        ;
multiplicative_expression
        : unary_expression
        | multiplicative_expression MUL unary_expression
        {
            $$ = &BinaryExpression{operator: MulOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        | multiplicative_expression DIV unary_expression
        {
            $$ = &BinaryExpression{operator: DivOperator, left: $1, right: $3}
            $$.SetPosition($1.Position())
        }
        ;
unary_expression
        : postfix_expression
        | SUB unary_expression
        {
            $$ = &MinusExpression{operand: $2}
            $$.SetPosition($1.Position())
        }
        | EXCLAMATION unary_expression
        {
            $$ = &LogicalNotExpression{operand: $2}
            $$.SetPosition($1.Position())
        }
        ;
postfix_expression
        : primary_expression
        ;
primary_expression
        : primary_no_new_array
        | IDENTIFIER
        {
            $$ = createIdentifierExpression($1.Lit, $1.Position());
        }
        ;
primary_no_new_array
        : primary_no_new_array LB expression RB
        {
            $$ = createIndexExpression($1, $3, $1.Position())
        }
        | IDENTIFIER LB expression RB
        {
            identifier := createIdentifierExpression($1.Lit, $1.Position());
            $$ = createIndexExpression(identifier, $3, $1.Position())
        }
        | primary_expression DOT IDENTIFIER
        {
            $$ = createMemberExpression($1, $3.Lit)
        }
        | primary_expression LP argument_list RP
        {
            $$ = &FunctionCallExpression{function: $1, argumentList: $3}
            $$.SetPosition($1.Position())
        }
        | primary_expression LP RP
        {
            $$ = &FunctionCallExpression{function: $1, argumentList: []Expression{}}
            $$.SetPosition($1.Position())
        }
        | LP expression RP
        {
            $$ = $2
        }
        | INT_LITERAL
        {
            value, _ := strconv.Atoi($1.Lit)
            $$ = &IntExpression{intValue: value}
            $$.SetPosition($1.Position())
        }
        | FLOAT_LITERAL
        {
            value, _ := strconv.ParseFloat($1.Lit, 64)
            $$ = &DoubleExpression{doubleValue: value}
            $$.SetPosition($1.Position())
        }
        | STRING_LITERAL
        {
            $$ = &StringExpression{stringValue: $1.Lit}
            $$.SetPosition($1.Position())
        }
        | TRUE_T
        {
            $$ = &BooleanExpression{booleanValue: true}
            $$.SetPosition($1.Position())
        }
        | FALSE_T
        {
            $$ = &BooleanExpression{booleanValue: false}
            $$.SetPosition($1.Position())
        }
        | NULL_T
        {
            $$ = &NullExpression{}
            $$.SetPosition($1.Position())
        }
        | array_literal
        ;
array_literal
        : LC expression_list RC
        {
            $$ = &ArrayLiteralExpression{arrayLiteral: $2}
            $$.SetPosition($1.Position())
        }
        | LC expression_list COMMA RC
        {
            $$ = &ArrayLiteralExpression{arrayLiteral: $2}
            $$.SetPosition($1.Position())
        }
        ;
expression_list
        :
        {
            $$ = nil
        }
        | assignment_expression
        {
            $$ = []Expression{$1}
        }
        | expression_list COMMA assignment_expression
        {
            $$ = append($1, $3)
        }
        ;
statement
        : expression SEMICOLON
        {
            $$ = &ExpressionStatement{expression: $1}
            $$.SetPosition($1.Position())
        }
        | if_statement
        | for_statement
        | return_statement
        | break_statement
        | continue_statement
        | declaration_statement
        ;
if_statement
        : IF expression block
        {
            $$ = &IfStatement{condition: $2, thenBlock: $3, elifList: []*ElseIf{}, elseBlock: nil}
            $$.SetPosition($1.Position())
        }
        | IF expression block ELSE block
        {
            $$ = &IfStatement{condition: $2, thenBlock: $3, elifList: []*ElseIf{}, elseBlock: $5}
            $$.SetPosition($1.Position())
        }
        | IF expression block else_if
        {
            $$ = &IfStatement{condition: $2, thenBlock: $3, elifList: $4, elseBlock: nil}
            $$.SetPosition($1.Position())
        }
        | IF expression block else_if ELSE block
        {
            $$ = &IfStatement{condition: $2, thenBlock: $3, elifList: $4, elseBlock: $6}
            $$.SetPosition($1.Position())
        }
        ;
else_if
        : ELSE IF expression block
        {
            $$ = []*ElseIf{&ElseIf{condition: $3, block: $4}}
        }
        | else_if ELSE IF expression block
        {
            $$ = append($1, &ElseIf{condition: $4, block: $5})
        }
        ;
for_statement
        : FOR LP expression_opt SEMICOLON expression_opt SEMICOLON expression_opt RP block
        {
            $$ = &ForStatement{init: $3, condition: $5, post: $7, block: $9}
            $$.SetPosition($1.Position())
            $9.parent = &StatementBlockInfo{statement: $$}
        }
        ;
expression_opt
        :
        {
            $$ = nil
        }
        | expression
        ;
return_statement
        : RETURN_T expression_opt SEMICOLON
        {
            $$ = &ReturnStatement{returnValue: $2};
            $$.SetPosition($1.Position())
        }
        ;
break_statement
        : BREAK SEMICOLON
        {
            $$ = &BreakStatement{}
            $$.SetPosition($1.Position())
        }
        ;
continue_statement
        : CONTINUE SEMICOLON
        {
            $$ = &ContinueStatement{}
            $$.SetPosition($1.Position())
        }
        ;
declaration_statement
        : VAR IDENTIFIER type_specifier SEMICOLON
        {
            $$ = &Declaration{typeSpecifier: $3, name: $2.Lit, variableIndex: -1}
            $$.SetPosition($1.Position())
        }
        | VAR IDENTIFIER type_specifier ASSIGN_T expression SEMICOLON
        {
            $$ = &Declaration{typeSpecifier: $3, name: $2.Lit, initializer: $5, variableIndex: -1}
            $$.SetPosition($1.Position())
        }
        ;
block
        : LC
        {
            l := yylex.(*Lexer)
            l.compiler.currentBlock = &Block{outerBlock: l.compiler.currentBlock}
            $<block>$ = l.compiler.currentBlock
        }
          statement_list RC
        {
            currentBlock := $<block>2
            currentBlock.statementList = $3

            l := yylex.(*Lexer)

            $<block>$ = l.compiler.currentBlock
            l.compiler.currentBlock = currentBlock.outerBlock
        }
        | LC RC
        {
            l := yylex.(*Lexer)
            $<block>$ = &Block{outerBlock: l.compiler.currentBlock}
        }
        ;
%%
