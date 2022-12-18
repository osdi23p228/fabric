// Code generated by counterfeiter. DO NOT EDIT.
package mock

import (
	"crypto/ecdsa"
	"sync"

	"github.com/osdi23p228/fabric/bccsp"
	"github.com/osdi23p228/fabric/bccsp/idemix/handlers"
)

type SignatureScheme struct {
	SignStub        func([]byte, handlers.Big, handlers.Ecp, handlers.Big, handlers.IssuerPublicKey, []bccsp.IdemixAttribute, []byte, int, []byte) ([]byte, error)
	signMutex       sync.RWMutex
	signArgsForCall []struct {
		arg1 []byte
		arg2 handlers.Big
		arg3 handlers.Ecp
		arg4 handlers.Big
		arg5 handlers.IssuerPublicKey
		arg6 []bccsp.IdemixAttribute
		arg7 []byte
		arg8 int
		arg9 []byte
	}
	signReturns struct {
		result1 []byte
		result2 error
	}
	signReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	VerifyStub        func(handlers.IssuerPublicKey, []byte, []byte, []bccsp.IdemixAttribute, int, *ecdsa.PublicKey, int) error
	verifyMutex       sync.RWMutex
	verifyArgsForCall []struct {
		arg1 handlers.IssuerPublicKey
		arg2 []byte
		arg3 []byte
		arg4 []bccsp.IdemixAttribute
		arg5 int
		arg6 *ecdsa.PublicKey
		arg7 int
	}
	verifyReturns struct {
		result1 error
	}
	verifyReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *SignatureScheme) Sign(arg1 []byte, arg2 handlers.Big, arg3 handlers.Ecp, arg4 handlers.Big, arg5 handlers.IssuerPublicKey, arg6 []bccsp.IdemixAttribute, arg7 []byte, arg8 int, arg9 []byte) ([]byte, error) {
	var arg1Copy []byte
	if arg1 != nil {
		arg1Copy = make([]byte, len(arg1))
		copy(arg1Copy, arg1)
	}
	var arg6Copy []bccsp.IdemixAttribute
	if arg6 != nil {
		arg6Copy = make([]bccsp.IdemixAttribute, len(arg6))
		copy(arg6Copy, arg6)
	}
	var arg7Copy []byte
	if arg7 != nil {
		arg7Copy = make([]byte, len(arg7))
		copy(arg7Copy, arg7)
	}
	var arg9Copy []byte
	if arg9 != nil {
		arg9Copy = make([]byte, len(arg9))
		copy(arg9Copy, arg9)
	}
	fake.signMutex.Lock()
	ret, specificReturn := fake.signReturnsOnCall[len(fake.signArgsForCall)]
	fake.signArgsForCall = append(fake.signArgsForCall, struct {
		arg1 []byte
		arg2 handlers.Big
		arg3 handlers.Ecp
		arg4 handlers.Big
		arg5 handlers.IssuerPublicKey
		arg6 []bccsp.IdemixAttribute
		arg7 []byte
		arg8 int
		arg9 []byte
	}{arg1Copy, arg2, arg3, arg4, arg5, arg6Copy, arg7Copy, arg8, arg9Copy})
	fake.recordInvocation("Sign", []interface{}{arg1Copy, arg2, arg3, arg4, arg5, arg6Copy, arg7Copy, arg8, arg9Copy})
	fake.signMutex.Unlock()
	if fake.SignStub != nil {
		return fake.SignStub(arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.signReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *SignatureScheme) SignCallCount() int {
	fake.signMutex.RLock()
	defer fake.signMutex.RUnlock()
	return len(fake.signArgsForCall)
}

func (fake *SignatureScheme) SignCalls(stub func([]byte, handlers.Big, handlers.Ecp, handlers.Big, handlers.IssuerPublicKey, []bccsp.IdemixAttribute, []byte, int, []byte) ([]byte, error)) {
	fake.signMutex.Lock()
	defer fake.signMutex.Unlock()
	fake.SignStub = stub
}

func (fake *SignatureScheme) SignArgsForCall(i int) ([]byte, handlers.Big, handlers.Ecp, handlers.Big, handlers.IssuerPublicKey, []bccsp.IdemixAttribute, []byte, int, []byte) {
	fake.signMutex.RLock()
	defer fake.signMutex.RUnlock()
	argsForCall := fake.signArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6, argsForCall.arg7, argsForCall.arg8, argsForCall.arg9
}

func (fake *SignatureScheme) SignReturns(result1 []byte, result2 error) {
	fake.signMutex.Lock()
	defer fake.signMutex.Unlock()
	fake.SignStub = nil
	fake.signReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *SignatureScheme) SignReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.signMutex.Lock()
	defer fake.signMutex.Unlock()
	fake.SignStub = nil
	if fake.signReturnsOnCall == nil {
		fake.signReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.signReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *SignatureScheme) Verify(arg1 handlers.IssuerPublicKey, arg2 []byte, arg3 []byte, arg4 []bccsp.IdemixAttribute, arg5 int, arg6 *ecdsa.PublicKey, arg7 int) error {
	var arg2Copy []byte
	if arg2 != nil {
		arg2Copy = make([]byte, len(arg2))
		copy(arg2Copy, arg2)
	}
	var arg3Copy []byte
	if arg3 != nil {
		arg3Copy = make([]byte, len(arg3))
		copy(arg3Copy, arg3)
	}
	var arg4Copy []bccsp.IdemixAttribute
	if arg4 != nil {
		arg4Copy = make([]bccsp.IdemixAttribute, len(arg4))
		copy(arg4Copy, arg4)
	}
	fake.verifyMutex.Lock()
	ret, specificReturn := fake.verifyReturnsOnCall[len(fake.verifyArgsForCall)]
	fake.verifyArgsForCall = append(fake.verifyArgsForCall, struct {
		arg1 handlers.IssuerPublicKey
		arg2 []byte
		arg3 []byte
		arg4 []bccsp.IdemixAttribute
		arg5 int
		arg6 *ecdsa.PublicKey
		arg7 int
	}{arg1, arg2Copy, arg3Copy, arg4Copy, arg5, arg6, arg7})
	fake.recordInvocation("Verify", []interface{}{arg1, arg2Copy, arg3Copy, arg4Copy, arg5, arg6, arg7})
	fake.verifyMutex.Unlock()
	if fake.VerifyStub != nil {
		return fake.VerifyStub(arg1, arg2, arg3, arg4, arg5, arg6, arg7)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.verifyReturns
	return fakeReturns.result1
}

func (fake *SignatureScheme) VerifyCallCount() int {
	fake.verifyMutex.RLock()
	defer fake.verifyMutex.RUnlock()
	return len(fake.verifyArgsForCall)
}

func (fake *SignatureScheme) VerifyCalls(stub func(handlers.IssuerPublicKey, []byte, []byte, []bccsp.IdemixAttribute, int, *ecdsa.PublicKey, int) error) {
	fake.verifyMutex.Lock()
	defer fake.verifyMutex.Unlock()
	fake.VerifyStub = stub
}

func (fake *SignatureScheme) VerifyArgsForCall(i int) (handlers.IssuerPublicKey, []byte, []byte, []bccsp.IdemixAttribute, int, *ecdsa.PublicKey, int) {
	fake.verifyMutex.RLock()
	defer fake.verifyMutex.RUnlock()
	argsForCall := fake.verifyArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4, argsForCall.arg5, argsForCall.arg6, argsForCall.arg7
}

func (fake *SignatureScheme) VerifyReturns(result1 error) {
	fake.verifyMutex.Lock()
	defer fake.verifyMutex.Unlock()
	fake.VerifyStub = nil
	fake.verifyReturns = struct {
		result1 error
	}{result1}
}

func (fake *SignatureScheme) VerifyReturnsOnCall(i int, result1 error) {
	fake.verifyMutex.Lock()
	defer fake.verifyMutex.Unlock()
	fake.VerifyStub = nil
	if fake.verifyReturnsOnCall == nil {
		fake.verifyReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.verifyReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *SignatureScheme) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.signMutex.RLock()
	defer fake.signMutex.RUnlock()
	fake.verifyMutex.RLock()
	defer fake.verifyMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *SignatureScheme) recordInvocation(key string, args []interface{}) {
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

var _ handlers.SignatureScheme = new(SignatureScheme)