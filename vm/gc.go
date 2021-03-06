package vm

var HeapThresholdSize = 1024 * 256

//////////////////////////////
// 垃圾回收
//////////////////////////////

//
// 判断是否下需要gc
//
func (vm *VirtualMachine) checkGC() {
	if len(vm.heap.objectList) > vm.heap.currentThreshold {
		vm.garbageCollect()

		vm.heap.currentThreshold += HeapThresholdSize
	}
}

//
// 标记，取消标记
//
func mark(ref *ObjectRef) {
	obj := ref.data
	if obj == nil {
		return
	}
	obj.setMark(true)

	switch o := obj.(type) {
	case *ObjectArrayObject:
		for _, subObj := range o.objectArray {
			mark(subObj)
		}
	case *ObjectClassObject:
		ec := ref.vTable.execClass

		for i, typ := range ec.fieldTypeList {
			if isReferenceType(typ) {
				mark(o.fieldList[i].(*ObjectRef))
			}
		}
	}
}

func resetMark(obj Object) {
	obj.setMark(false)
}

//
// 标记
//
// TODO
func (vm *VirtualMachine) markObjects() {
	for _, obj := range vm.heap.objectList {
		resetMark(obj)
	}

	for _, ee := range vm.executableEntryList {
		for i, variable := range ee.static.variableList {
			if isReferenceType(ee.executable.GlobalVariableList[i].typeSpecifier) {
				mark(variable.(*ObjectRef))
			}
		}
	}

	for i := 0; i < vm.stack.stackPointer; i++ {
		if vm.stack.stack[i].isPointer() {
			o := vm.stack.stack[i].(*ObjectRef)
			mark(o)
		}
	}
}

//
// 删除对象
//
func (vm *VirtualMachine) disposeObject(obj Object) {
	switch o := obj.(type) {
	case *ObjectString:
		//
	case *ObjectArrayInt:
		o.intArray = nil
	case *ObjectArrayDouble:
		o.doubleArray = nil
	case *ObjectArrayObject:
		o.objectArray = nil
	default:
		panic("TODO")
	}

	obj = nil
}

//
// 清理
//
func (vm *VirtualMachine) sweepObjects() {
	newObjectList := []Object{}
	for _, obj := range vm.heap.objectList {
		if !obj.isMarked() {
			vm.disposeObject(obj)
		} else {
			newObjectList = append(newObjectList, obj)
		}
	}
	vm.heap.objectList = newObjectList
}

func (vm *VirtualMachine) garbageCollect() {
	vm.markObjects()
	vm.sweepObjects()
}

//////////////////////////////
// 创建对象
//////////////////////////////

//
// 添加对象到堆, 用于垃圾回收
//
func (vm *VirtualMachine) addObject(value Object) {
	vm.checkGC()
	value.setMark(false)
	vm.heap.append(value)
}

//////////////////////////////

//
// string object
//
func (vm *VirtualMachine) createStringObject(str string) *ObjectRef {
	ret := &ObjectString{}
	vm.addObject(ret)

	ret.stringValue = str

	ref := &ObjectRef{data: ret}

	return ref
}

//
// Array object
//
func (vm *VirtualMachine) createArrayInt(size int) *ObjectRef {
	obj := &ObjectArrayInt{intArray: make([]int, size)}
	vm.addObject(obj)

	ref := &ObjectRef{data: obj}

	return ref
}

func (vm *VirtualMachine) createArrayDouble(size int) *ObjectRef {
	obj := &ObjectArrayDouble{doubleArray: make([]float64, size)}
	vm.addObject(obj)

	ref := &ObjectRef{data: obj}

	return ref
}

func (vm *VirtualMachine) createArrayObject(size int) *ObjectRef {
	obj := &ObjectArrayObject{objectArray: make([]*ObjectRef, size)}
	vm.addObject(obj)

	ref := &ObjectRef{data: obj}

	return ref
}

//
// class object
//
func (vm *VirtualMachine) createClassObject(classIndex int) *ObjectRef {
	obj := &ObjectClassObject{}
	vm.addObject(obj)

	execClass := vm.classList[classIndex]

	obj.fieldList = []Value{}
	for _, typ := range execClass.fieldTypeList {
		obj.fieldList = append(obj.fieldList, initializeValue(typ))
	}

	ref := &ObjectRef{
		vTable: execClass.classTable,
		data:   obj,
	}

	return ref
}

// utils

// 判断是否是引用类型
func isReferenceType(typ *TypeSpecifier) bool {
	// 字符串, 类, 数组
	if ((typ.BasicType == StringType || typ.BasicType == ClassType) && len(typ.DeriveList) == 0) || (typ.isArrayDerive()) {
		return true
	}
	return false
}

// 连接字符对象
func (vm *VirtualMachine) chainStringObject(str1, str2 *ObjectRef) *ObjectRef {
	var left, right string
	if str1.data == nil {
		left = "null"
	} else {
		left = str1.data.(*ObjectString).stringValue
	}

	if str2.data == nil {
		right = "null"
	} else {
		right = str2.data.(*ObjectString).stringValue
	}

	str := left + right
	ret := vm.createStringObject(str)
	return ret
}
