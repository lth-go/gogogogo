package vm

const (
	FUNCTION_NOT_FOUND_ERR int = iota
	FUNCTION_MULTIPLE_DEFINE_ERR
	INDEX_OUT_OF_BOUNDS_ERR
	DIVISION_BY_ZERO_ERR
	NULL_POINTER_ERR
	LOAD_FILE_NOT_FOUND_ERR
	LOAD_FILE_ERR
	CLASS_MULTIPLE_DEFINE_ERR
	CLASS_NOT_FOUND_ERR
	CLASS_CAST_ERR
	DYNAMIC_LOAD_WITHOUT_PACKAGE_ERR
)