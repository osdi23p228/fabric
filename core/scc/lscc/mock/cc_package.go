// Code generated by counterfeiter. DO NOT EDIT.
package mock

import (
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/osdi23p228/fabric/core/common/ccprovider"
)

type CCPackage struct {
	GetChaincodeDataStub        func() *ccprovider.ChaincodeData
	getChaincodeDataMutex       sync.RWMutex
	getChaincodeDataArgsForCall []struct {
	}
	getChaincodeDataReturns struct {
		result1 *ccprovider.ChaincodeData
	}
	getChaincodeDataReturnsOnCall map[int]struct {
		result1 *ccprovider.ChaincodeData
	}
	GetDepSpecStub        func() *peer.ChaincodeDeploymentSpec
	getDepSpecMutex       sync.RWMutex
	getDepSpecArgsForCall []struct {
	}
	getDepSpecReturns struct {
		result1 *peer.ChaincodeDeploymentSpec
	}
	getDepSpecReturnsOnCall map[int]struct {
		result1 *peer.ChaincodeDeploymentSpec
	}
	GetDepSpecBytesStub        func() []byte
	getDepSpecBytesMutex       sync.RWMutex
	getDepSpecBytesArgsForCall []struct {
	}
	getDepSpecBytesReturns struct {
		result1 []byte
	}
	getDepSpecBytesReturnsOnCall map[int]struct {
		result1 []byte
	}
	GetIdStub        func() []byte
	getIdMutex       sync.RWMutex
	getIdArgsForCall []struct {
	}
	getIdReturns struct {
		result1 []byte
	}
	getIdReturnsOnCall map[int]struct {
		result1 []byte
	}
	GetPackageObjectStub        func() proto.Message
	getPackageObjectMutex       sync.RWMutex
	getPackageObjectArgsForCall []struct {
	}
	getPackageObjectReturns struct {
		result1 proto.Message
	}
	getPackageObjectReturnsOnCall map[int]struct {
		result1 proto.Message
	}
	InitFromBufferStub        func([]byte) (*ccprovider.ChaincodeData, error)
	initFromBufferMutex       sync.RWMutex
	initFromBufferArgsForCall []struct {
		arg1 []byte
	}
	initFromBufferReturns struct {
		result1 *ccprovider.ChaincodeData
		result2 error
	}
	initFromBufferReturnsOnCall map[int]struct {
		result1 *ccprovider.ChaincodeData
		result2 error
	}
	PutChaincodeToFSStub        func() error
	putChaincodeToFSMutex       sync.RWMutex
	putChaincodeToFSArgsForCall []struct {
	}
	putChaincodeToFSReturns struct {
		result1 error
	}
	putChaincodeToFSReturnsOnCall map[int]struct {
		result1 error
	}
	ValidateCCStub        func(*ccprovider.ChaincodeData) error
	validateCCMutex       sync.RWMutex
	validateCCArgsForCall []struct {
		arg1 *ccprovider.ChaincodeData
	}
	validateCCReturns struct {
		result1 error
	}
	validateCCReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *CCPackage) GetChaincodeData() *ccprovider.ChaincodeData {
	fake.getChaincodeDataMutex.Lock()
	ret, specificReturn := fake.getChaincodeDataReturnsOnCall[len(fake.getChaincodeDataArgsForCall)]
	fake.getChaincodeDataArgsForCall = append(fake.getChaincodeDataArgsForCall, struct {
	}{})
	fake.recordInvocation("GetChaincodeData", []interface{}{})
	fake.getChaincodeDataMutex.Unlock()
	if fake.GetChaincodeDataStub != nil {
		return fake.GetChaincodeDataStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getChaincodeDataReturns
	return fakeReturns.result1
}

func (fake *CCPackage) GetChaincodeDataCallCount() int {
	fake.getChaincodeDataMutex.RLock()
	defer fake.getChaincodeDataMutex.RUnlock()
	return len(fake.getChaincodeDataArgsForCall)
}

func (fake *CCPackage) GetChaincodeDataCalls(stub func() *ccprovider.ChaincodeData) {
	fake.getChaincodeDataMutex.Lock()
	defer fake.getChaincodeDataMutex.Unlock()
	fake.GetChaincodeDataStub = stub
}

func (fake *CCPackage) GetChaincodeDataReturns(result1 *ccprovider.ChaincodeData) {
	fake.getChaincodeDataMutex.Lock()
	defer fake.getChaincodeDataMutex.Unlock()
	fake.GetChaincodeDataStub = nil
	fake.getChaincodeDataReturns = struct {
		result1 *ccprovider.ChaincodeData
	}{result1}
}

func (fake *CCPackage) GetChaincodeDataReturnsOnCall(i int, result1 *ccprovider.ChaincodeData) {
	fake.getChaincodeDataMutex.Lock()
	defer fake.getChaincodeDataMutex.Unlock()
	fake.GetChaincodeDataStub = nil
	if fake.getChaincodeDataReturnsOnCall == nil {
		fake.getChaincodeDataReturnsOnCall = make(map[int]struct {
			result1 *ccprovider.ChaincodeData
		})
	}
	fake.getChaincodeDataReturnsOnCall[i] = struct {
		result1 *ccprovider.ChaincodeData
	}{result1}
}

func (fake *CCPackage) GetDepSpec() *peer.ChaincodeDeploymentSpec {
	fake.getDepSpecMutex.Lock()
	ret, specificReturn := fake.getDepSpecReturnsOnCall[len(fake.getDepSpecArgsForCall)]
	fake.getDepSpecArgsForCall = append(fake.getDepSpecArgsForCall, struct {
	}{})
	fake.recordInvocation("GetDepSpec", []interface{}{})
	fake.getDepSpecMutex.Unlock()
	if fake.GetDepSpecStub != nil {
		return fake.GetDepSpecStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getDepSpecReturns
	return fakeReturns.result1
}

func (fake *CCPackage) GetDepSpecCallCount() int {
	fake.getDepSpecMutex.RLock()
	defer fake.getDepSpecMutex.RUnlock()
	return len(fake.getDepSpecArgsForCall)
}

func (fake *CCPackage) GetDepSpecCalls(stub func() *peer.ChaincodeDeploymentSpec) {
	fake.getDepSpecMutex.Lock()
	defer fake.getDepSpecMutex.Unlock()
	fake.GetDepSpecStub = stub
}

func (fake *CCPackage) GetDepSpecReturns(result1 *peer.ChaincodeDeploymentSpec) {
	fake.getDepSpecMutex.Lock()
	defer fake.getDepSpecMutex.Unlock()
	fake.GetDepSpecStub = nil
	fake.getDepSpecReturns = struct {
		result1 *peer.ChaincodeDeploymentSpec
	}{result1}
}

func (fake *CCPackage) GetDepSpecReturnsOnCall(i int, result1 *peer.ChaincodeDeploymentSpec) {
	fake.getDepSpecMutex.Lock()
	defer fake.getDepSpecMutex.Unlock()
	fake.GetDepSpecStub = nil
	if fake.getDepSpecReturnsOnCall == nil {
		fake.getDepSpecReturnsOnCall = make(map[int]struct {
			result1 *peer.ChaincodeDeploymentSpec
		})
	}
	fake.getDepSpecReturnsOnCall[i] = struct {
		result1 *peer.ChaincodeDeploymentSpec
	}{result1}
}

func (fake *CCPackage) GetDepSpecBytes() []byte {
	fake.getDepSpecBytesMutex.Lock()
	ret, specificReturn := fake.getDepSpecBytesReturnsOnCall[len(fake.getDepSpecBytesArgsForCall)]
	fake.getDepSpecBytesArgsForCall = append(fake.getDepSpecBytesArgsForCall, struct {
	}{})
	fake.recordInvocation("GetDepSpecBytes", []interface{}{})
	fake.getDepSpecBytesMutex.Unlock()
	if fake.GetDepSpecBytesStub != nil {
		return fake.GetDepSpecBytesStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getDepSpecBytesReturns
	return fakeReturns.result1
}

func (fake *CCPackage) GetDepSpecBytesCallCount() int {
	fake.getDepSpecBytesMutex.RLock()
	defer fake.getDepSpecBytesMutex.RUnlock()
	return len(fake.getDepSpecBytesArgsForCall)
}

func (fake *CCPackage) GetDepSpecBytesCalls(stub func() []byte) {
	fake.getDepSpecBytesMutex.Lock()
	defer fake.getDepSpecBytesMutex.Unlock()
	fake.GetDepSpecBytesStub = stub
}

func (fake *CCPackage) GetDepSpecBytesReturns(result1 []byte) {
	fake.getDepSpecBytesMutex.Lock()
	defer fake.getDepSpecBytesMutex.Unlock()
	fake.GetDepSpecBytesStub = nil
	fake.getDepSpecBytesReturns = struct {
		result1 []byte
	}{result1}
}

func (fake *CCPackage) GetDepSpecBytesReturnsOnCall(i int, result1 []byte) {
	fake.getDepSpecBytesMutex.Lock()
	defer fake.getDepSpecBytesMutex.Unlock()
	fake.GetDepSpecBytesStub = nil
	if fake.getDepSpecBytesReturnsOnCall == nil {
		fake.getDepSpecBytesReturnsOnCall = make(map[int]struct {
			result1 []byte
		})
	}
	fake.getDepSpecBytesReturnsOnCall[i] = struct {
		result1 []byte
	}{result1}
}

func (fake *CCPackage) GetId() []byte {
	fake.getIdMutex.Lock()
	ret, specificReturn := fake.getIdReturnsOnCall[len(fake.getIdArgsForCall)]
	fake.getIdArgsForCall = append(fake.getIdArgsForCall, struct {
	}{})
	fake.recordInvocation("GetId", []interface{}{})
	fake.getIdMutex.Unlock()
	if fake.GetIdStub != nil {
		return fake.GetIdStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getIdReturns
	return fakeReturns.result1
}

func (fake *CCPackage) GetIdCallCount() int {
	fake.getIdMutex.RLock()
	defer fake.getIdMutex.RUnlock()
	return len(fake.getIdArgsForCall)
}

func (fake *CCPackage) GetIdCalls(stub func() []byte) {
	fake.getIdMutex.Lock()
	defer fake.getIdMutex.Unlock()
	fake.GetIdStub = stub
}

func (fake *CCPackage) GetIdReturns(result1 []byte) {
	fake.getIdMutex.Lock()
	defer fake.getIdMutex.Unlock()
	fake.GetIdStub = nil
	fake.getIdReturns = struct {
		result1 []byte
	}{result1}
}

func (fake *CCPackage) GetIdReturnsOnCall(i int, result1 []byte) {
	fake.getIdMutex.Lock()
	defer fake.getIdMutex.Unlock()
	fake.GetIdStub = nil
	if fake.getIdReturnsOnCall == nil {
		fake.getIdReturnsOnCall = make(map[int]struct {
			result1 []byte
		})
	}
	fake.getIdReturnsOnCall[i] = struct {
		result1 []byte
	}{result1}
}

func (fake *CCPackage) GetPackageObject() proto.Message {
	fake.getPackageObjectMutex.Lock()
	ret, specificReturn := fake.getPackageObjectReturnsOnCall[len(fake.getPackageObjectArgsForCall)]
	fake.getPackageObjectArgsForCall = append(fake.getPackageObjectArgsForCall, struct {
	}{})
	fake.recordInvocation("GetPackageObject", []interface{}{})
	fake.getPackageObjectMutex.Unlock()
	if fake.GetPackageObjectStub != nil {
		return fake.GetPackageObjectStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.getPackageObjectReturns
	return fakeReturns.result1
}

func (fake *CCPackage) GetPackageObjectCallCount() int {
	fake.getPackageObjectMutex.RLock()
	defer fake.getPackageObjectMutex.RUnlock()
	return len(fake.getPackageObjectArgsForCall)
}

func (fake *CCPackage) GetPackageObjectCalls(stub func() proto.Message) {
	fake.getPackageObjectMutex.Lock()
	defer fake.getPackageObjectMutex.Unlock()
	fake.GetPackageObjectStub = stub
}

func (fake *CCPackage) GetPackageObjectReturns(result1 proto.Message) {
	fake.getPackageObjectMutex.Lock()
	defer fake.getPackageObjectMutex.Unlock()
	fake.GetPackageObjectStub = nil
	fake.getPackageObjectReturns = struct {
		result1 proto.Message
	}{result1}
}

func (fake *CCPackage) GetPackageObjectReturnsOnCall(i int, result1 proto.Message) {
	fake.getPackageObjectMutex.Lock()
	defer fake.getPackageObjectMutex.Unlock()
	fake.GetPackageObjectStub = nil
	if fake.getPackageObjectReturnsOnCall == nil {
		fake.getPackageObjectReturnsOnCall = make(map[int]struct {
			result1 proto.Message
		})
	}
	fake.getPackageObjectReturnsOnCall[i] = struct {
		result1 proto.Message
	}{result1}
}

func (fake *CCPackage) InitFromBuffer(arg1 []byte) (*ccprovider.ChaincodeData, error) {
	var arg1Copy []byte
	if arg1 != nil {
		arg1Copy = make([]byte, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.initFromBufferMutex.Lock()
	ret, specificReturn := fake.initFromBufferReturnsOnCall[len(fake.initFromBufferArgsForCall)]
	fake.initFromBufferArgsForCall = append(fake.initFromBufferArgsForCall, struct {
		arg1 []byte
	}{arg1Copy})
	fake.recordInvocation("InitFromBuffer", []interface{}{arg1Copy})
	fake.initFromBufferMutex.Unlock()
	if fake.InitFromBufferStub != nil {
		return fake.InitFromBufferStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.initFromBufferReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *CCPackage) InitFromBufferCallCount() int {
	fake.initFromBufferMutex.RLock()
	defer fake.initFromBufferMutex.RUnlock()
	return len(fake.initFromBufferArgsForCall)
}

func (fake *CCPackage) InitFromBufferCalls(stub func([]byte) (*ccprovider.ChaincodeData, error)) {
	fake.initFromBufferMutex.Lock()
	defer fake.initFromBufferMutex.Unlock()
	fake.InitFromBufferStub = stub
}

func (fake *CCPackage) InitFromBufferArgsForCall(i int) []byte {
	fake.initFromBufferMutex.RLock()
	defer fake.initFromBufferMutex.RUnlock()
	argsForCall := fake.initFromBufferArgsForCall[i]
	return argsForCall.arg1
}

func (fake *CCPackage) InitFromBufferReturns(result1 *ccprovider.ChaincodeData, result2 error) {
	fake.initFromBufferMutex.Lock()
	defer fake.initFromBufferMutex.Unlock()
	fake.InitFromBufferStub = nil
	fake.initFromBufferReturns = struct {
		result1 *ccprovider.ChaincodeData
		result2 error
	}{result1, result2}
}

func (fake *CCPackage) InitFromBufferReturnsOnCall(i int, result1 *ccprovider.ChaincodeData, result2 error) {
	fake.initFromBufferMutex.Lock()
	defer fake.initFromBufferMutex.Unlock()
	fake.InitFromBufferStub = nil
	if fake.initFromBufferReturnsOnCall == nil {
		fake.initFromBufferReturnsOnCall = make(map[int]struct {
			result1 *ccprovider.ChaincodeData
			result2 error
		})
	}
	fake.initFromBufferReturnsOnCall[i] = struct {
		result1 *ccprovider.ChaincodeData
		result2 error
	}{result1, result2}
}

func (fake *CCPackage) PutChaincodeToFS() error {
	fake.putChaincodeToFSMutex.Lock()
	ret, specificReturn := fake.putChaincodeToFSReturnsOnCall[len(fake.putChaincodeToFSArgsForCall)]
	fake.putChaincodeToFSArgsForCall = append(fake.putChaincodeToFSArgsForCall, struct {
	}{})
	fake.recordInvocation("PutChaincodeToFS", []interface{}{})
	fake.putChaincodeToFSMutex.Unlock()
	if fake.PutChaincodeToFSStub != nil {
		return fake.PutChaincodeToFSStub()
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.putChaincodeToFSReturns
	return fakeReturns.result1
}

func (fake *CCPackage) PutChaincodeToFSCallCount() int {
	fake.putChaincodeToFSMutex.RLock()
	defer fake.putChaincodeToFSMutex.RUnlock()
	return len(fake.putChaincodeToFSArgsForCall)
}

func (fake *CCPackage) PutChaincodeToFSCalls(stub func() error) {
	fake.putChaincodeToFSMutex.Lock()
	defer fake.putChaincodeToFSMutex.Unlock()
	fake.PutChaincodeToFSStub = stub
}

func (fake *CCPackage) PutChaincodeToFSReturns(result1 error) {
	fake.putChaincodeToFSMutex.Lock()
	defer fake.putChaincodeToFSMutex.Unlock()
	fake.PutChaincodeToFSStub = nil
	fake.putChaincodeToFSReturns = struct {
		result1 error
	}{result1}
}

func (fake *CCPackage) PutChaincodeToFSReturnsOnCall(i int, result1 error) {
	fake.putChaincodeToFSMutex.Lock()
	defer fake.putChaincodeToFSMutex.Unlock()
	fake.PutChaincodeToFSStub = nil
	if fake.putChaincodeToFSReturnsOnCall == nil {
		fake.putChaincodeToFSReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.putChaincodeToFSReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *CCPackage) ValidateCC(arg1 *ccprovider.ChaincodeData) error {
	fake.validateCCMutex.Lock()
	ret, specificReturn := fake.validateCCReturnsOnCall[len(fake.validateCCArgsForCall)]
	fake.validateCCArgsForCall = append(fake.validateCCArgsForCall, struct {
		arg1 *ccprovider.ChaincodeData
	}{arg1})
	fake.recordInvocation("ValidateCC", []interface{}{arg1})
	fake.validateCCMutex.Unlock()
	if fake.ValidateCCStub != nil {
		return fake.ValidateCCStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.validateCCReturns
	return fakeReturns.result1
}

func (fake *CCPackage) ValidateCCCallCount() int {
	fake.validateCCMutex.RLock()
	defer fake.validateCCMutex.RUnlock()
	return len(fake.validateCCArgsForCall)
}

func (fake *CCPackage) ValidateCCCalls(stub func(*ccprovider.ChaincodeData) error) {
	fake.validateCCMutex.Lock()
	defer fake.validateCCMutex.Unlock()
	fake.ValidateCCStub = stub
}

func (fake *CCPackage) ValidateCCArgsForCall(i int) *ccprovider.ChaincodeData {
	fake.validateCCMutex.RLock()
	defer fake.validateCCMutex.RUnlock()
	argsForCall := fake.validateCCArgsForCall[i]
	return argsForCall.arg1
}

func (fake *CCPackage) ValidateCCReturns(result1 error) {
	fake.validateCCMutex.Lock()
	defer fake.validateCCMutex.Unlock()
	fake.ValidateCCStub = nil
	fake.validateCCReturns = struct {
		result1 error
	}{result1}
}

func (fake *CCPackage) ValidateCCReturnsOnCall(i int, result1 error) {
	fake.validateCCMutex.Lock()
	defer fake.validateCCMutex.Unlock()
	fake.ValidateCCStub = nil
	if fake.validateCCReturnsOnCall == nil {
		fake.validateCCReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.validateCCReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *CCPackage) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getChaincodeDataMutex.RLock()
	defer fake.getChaincodeDataMutex.RUnlock()
	fake.getDepSpecMutex.RLock()
	defer fake.getDepSpecMutex.RUnlock()
	fake.getDepSpecBytesMutex.RLock()
	defer fake.getDepSpecBytesMutex.RUnlock()
	fake.getIdMutex.RLock()
	defer fake.getIdMutex.RUnlock()
	fake.getPackageObjectMutex.RLock()
	defer fake.getPackageObjectMutex.RUnlock()
	fake.initFromBufferMutex.RLock()
	defer fake.initFromBufferMutex.RUnlock()
	fake.putChaincodeToFSMutex.RLock()
	defer fake.putChaincodeToFSMutex.RUnlock()
	fake.validateCCMutex.RLock()
	defer fake.validateCCMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *CCPackage) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}