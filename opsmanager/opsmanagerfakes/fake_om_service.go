// Code generated by counterfeiter. DO NOT EDIT.
package opsmanagerfakes

import (
	"io"
	"sync"

	"github.com/pivotal-cf/aqueduct-courier/opsmanager"
)

type FakeOmService struct {
	CertificateAuthoritiesStub        func() (io.Reader, error)
	certificateAuthoritiesMutex       sync.RWMutex
	certificateAuthoritiesArgsForCall []struct {
	}
	certificateAuthoritiesReturns struct {
		result1 io.Reader
		result2 error
	}
	certificateAuthoritiesReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	CertificatesStub        func() (io.Reader, error)
	certificatesMutex       sync.RWMutex
	certificatesArgsForCall []struct {
	}
	certificatesReturns struct {
		result1 io.Reader
		result2 error
	}
	certificatesReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	DeployedProductsStub        func() (io.Reader, error)
	deployedProductsMutex       sync.RWMutex
	deployedProductsArgsForCall []struct {
	}
	deployedProductsReturns struct {
		result1 io.Reader
		result2 error
	}
	deployedProductsReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	DiagnosticReportStub        func() (io.Reader, error)
	diagnosticReportMutex       sync.RWMutex
	diagnosticReportArgsForCall []struct {
	}
	diagnosticReportReturns struct {
		result1 io.Reader
		result2 error
	}
	diagnosticReportReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	InstallationsStub        func() (io.Reader, error)
	installationsMutex       sync.RWMutex
	installationsArgsForCall []struct {
	}
	installationsReturns struct {
		result1 io.Reader
		result2 error
	}
	installationsReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	PendingChangesStub        func() (io.Reader, error)
	pendingChangesMutex       sync.RWMutex
	pendingChangesArgsForCall []struct {
	}
	pendingChangesReturns struct {
		result1 io.Reader
		result2 error
	}
	pendingChangesReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	ProductPropertiesStub        func(string) (io.Reader, error)
	productPropertiesMutex       sync.RWMutex
	productPropertiesArgsForCall []struct {
		arg1 string
	}
	productPropertiesReturns struct {
		result1 io.Reader
		result2 error
	}
	productPropertiesReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	ProductResourcesStub        func(string) (io.Reader, error)
	productResourcesMutex       sync.RWMutex
	productResourcesArgsForCall []struct {
		arg1 string
	}
	productResourcesReturns struct {
		result1 io.Reader
		result2 error
	}
	productResourcesReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	VmTypesStub        func() (io.Reader, error)
	vmTypesMutex       sync.RWMutex
	vmTypesArgsForCall []struct {
	}
	vmTypesReturns struct {
		result1 io.Reader
		result2 error
	}
	vmTypesReturnsOnCall map[int]struct {
		result1 io.Reader
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeOmService) CertificateAuthorities() (io.Reader, error) {
	fake.certificateAuthoritiesMutex.Lock()
	ret, specificReturn := fake.certificateAuthoritiesReturnsOnCall[len(fake.certificateAuthoritiesArgsForCall)]
	fake.certificateAuthoritiesArgsForCall = append(fake.certificateAuthoritiesArgsForCall, struct {
	}{})
	stub := fake.CertificateAuthoritiesStub
	fakeReturns := fake.certificateAuthoritiesReturns
	fake.recordInvocation("CertificateAuthorities", []interface{}{})
	fake.certificateAuthoritiesMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOmService) CertificateAuthoritiesCallCount() int {
	fake.certificateAuthoritiesMutex.RLock()
	defer fake.certificateAuthoritiesMutex.RUnlock()
	return len(fake.certificateAuthoritiesArgsForCall)
}

func (fake *FakeOmService) CertificateAuthoritiesCalls(stub func() (io.Reader, error)) {
	fake.certificateAuthoritiesMutex.Lock()
	defer fake.certificateAuthoritiesMutex.Unlock()
	fake.CertificateAuthoritiesStub = stub
}

func (fake *FakeOmService) CertificateAuthoritiesReturns(result1 io.Reader, result2 error) {
	fake.certificateAuthoritiesMutex.Lock()
	defer fake.certificateAuthoritiesMutex.Unlock()
	fake.CertificateAuthoritiesStub = nil
	fake.certificateAuthoritiesReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) CertificateAuthoritiesReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.certificateAuthoritiesMutex.Lock()
	defer fake.certificateAuthoritiesMutex.Unlock()
	fake.CertificateAuthoritiesStub = nil
	if fake.certificateAuthoritiesReturnsOnCall == nil {
		fake.certificateAuthoritiesReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.certificateAuthoritiesReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) Certificates() (io.Reader, error) {
	fake.certificatesMutex.Lock()
	ret, specificReturn := fake.certificatesReturnsOnCall[len(fake.certificatesArgsForCall)]
	fake.certificatesArgsForCall = append(fake.certificatesArgsForCall, struct {
	}{})
	stub := fake.CertificatesStub
	fakeReturns := fake.certificatesReturns
	fake.recordInvocation("Certificates", []interface{}{})
	fake.certificatesMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOmService) CertificatesCallCount() int {
	fake.certificatesMutex.RLock()
	defer fake.certificatesMutex.RUnlock()
	return len(fake.certificatesArgsForCall)
}

func (fake *FakeOmService) CertificatesCalls(stub func() (io.Reader, error)) {
	fake.certificatesMutex.Lock()
	defer fake.certificatesMutex.Unlock()
	fake.CertificatesStub = stub
}

func (fake *FakeOmService) CertificatesReturns(result1 io.Reader, result2 error) {
	fake.certificatesMutex.Lock()
	defer fake.certificatesMutex.Unlock()
	fake.CertificatesStub = nil
	fake.certificatesReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) CertificatesReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.certificatesMutex.Lock()
	defer fake.certificatesMutex.Unlock()
	fake.CertificatesStub = nil
	if fake.certificatesReturnsOnCall == nil {
		fake.certificatesReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.certificatesReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) DeployedProducts() (io.Reader, error) {
	fake.deployedProductsMutex.Lock()
	ret, specificReturn := fake.deployedProductsReturnsOnCall[len(fake.deployedProductsArgsForCall)]
	fake.deployedProductsArgsForCall = append(fake.deployedProductsArgsForCall, struct {
	}{})
	stub := fake.DeployedProductsStub
	fakeReturns := fake.deployedProductsReturns
	fake.recordInvocation("DeployedProducts", []interface{}{})
	fake.deployedProductsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOmService) DeployedProductsCallCount() int {
	fake.deployedProductsMutex.RLock()
	defer fake.deployedProductsMutex.RUnlock()
	return len(fake.deployedProductsArgsForCall)
}

func (fake *FakeOmService) DeployedProductsCalls(stub func() (io.Reader, error)) {
	fake.deployedProductsMutex.Lock()
	defer fake.deployedProductsMutex.Unlock()
	fake.DeployedProductsStub = stub
}

func (fake *FakeOmService) DeployedProductsReturns(result1 io.Reader, result2 error) {
	fake.deployedProductsMutex.Lock()
	defer fake.deployedProductsMutex.Unlock()
	fake.DeployedProductsStub = nil
	fake.deployedProductsReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) DeployedProductsReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.deployedProductsMutex.Lock()
	defer fake.deployedProductsMutex.Unlock()
	fake.DeployedProductsStub = nil
	if fake.deployedProductsReturnsOnCall == nil {
		fake.deployedProductsReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.deployedProductsReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) DiagnosticReport() (io.Reader, error) {
	fake.diagnosticReportMutex.Lock()
	ret, specificReturn := fake.diagnosticReportReturnsOnCall[len(fake.diagnosticReportArgsForCall)]
	fake.diagnosticReportArgsForCall = append(fake.diagnosticReportArgsForCall, struct {
	}{})
	stub := fake.DiagnosticReportStub
	fakeReturns := fake.diagnosticReportReturns
	fake.recordInvocation("DiagnosticReport", []interface{}{})
	fake.diagnosticReportMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOmService) DiagnosticReportCallCount() int {
	fake.diagnosticReportMutex.RLock()
	defer fake.diagnosticReportMutex.RUnlock()
	return len(fake.diagnosticReportArgsForCall)
}

func (fake *FakeOmService) DiagnosticReportCalls(stub func() (io.Reader, error)) {
	fake.diagnosticReportMutex.Lock()
	defer fake.diagnosticReportMutex.Unlock()
	fake.DiagnosticReportStub = stub
}

func (fake *FakeOmService) DiagnosticReportReturns(result1 io.Reader, result2 error) {
	fake.diagnosticReportMutex.Lock()
	defer fake.diagnosticReportMutex.Unlock()
	fake.DiagnosticReportStub = nil
	fake.diagnosticReportReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) DiagnosticReportReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.diagnosticReportMutex.Lock()
	defer fake.diagnosticReportMutex.Unlock()
	fake.DiagnosticReportStub = nil
	if fake.diagnosticReportReturnsOnCall == nil {
		fake.diagnosticReportReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.diagnosticReportReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) Installations() (io.Reader, error) {
	fake.installationsMutex.Lock()
	ret, specificReturn := fake.installationsReturnsOnCall[len(fake.installationsArgsForCall)]
	fake.installationsArgsForCall = append(fake.installationsArgsForCall, struct {
	}{})
	stub := fake.InstallationsStub
	fakeReturns := fake.installationsReturns
	fake.recordInvocation("Installations", []interface{}{})
	fake.installationsMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOmService) InstallationsCallCount() int {
	fake.installationsMutex.RLock()
	defer fake.installationsMutex.RUnlock()
	return len(fake.installationsArgsForCall)
}

func (fake *FakeOmService) InstallationsCalls(stub func() (io.Reader, error)) {
	fake.installationsMutex.Lock()
	defer fake.installationsMutex.Unlock()
	fake.InstallationsStub = stub
}

func (fake *FakeOmService) InstallationsReturns(result1 io.Reader, result2 error) {
	fake.installationsMutex.Lock()
	defer fake.installationsMutex.Unlock()
	fake.InstallationsStub = nil
	fake.installationsReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) InstallationsReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.installationsMutex.Lock()
	defer fake.installationsMutex.Unlock()
	fake.InstallationsStub = nil
	if fake.installationsReturnsOnCall == nil {
		fake.installationsReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.installationsReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) PendingChanges() (io.Reader, error) {
	fake.pendingChangesMutex.Lock()
	ret, specificReturn := fake.pendingChangesReturnsOnCall[len(fake.pendingChangesArgsForCall)]
	fake.pendingChangesArgsForCall = append(fake.pendingChangesArgsForCall, struct {
	}{})
	stub := fake.PendingChangesStub
	fakeReturns := fake.pendingChangesReturns
	fake.recordInvocation("PendingChanges", []interface{}{})
	fake.pendingChangesMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOmService) PendingChangesCallCount() int {
	fake.pendingChangesMutex.RLock()
	defer fake.pendingChangesMutex.RUnlock()
	return len(fake.pendingChangesArgsForCall)
}

func (fake *FakeOmService) PendingChangesCalls(stub func() (io.Reader, error)) {
	fake.pendingChangesMutex.Lock()
	defer fake.pendingChangesMutex.Unlock()
	fake.PendingChangesStub = stub
}

func (fake *FakeOmService) PendingChangesReturns(result1 io.Reader, result2 error) {
	fake.pendingChangesMutex.Lock()
	defer fake.pendingChangesMutex.Unlock()
	fake.PendingChangesStub = nil
	fake.pendingChangesReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) PendingChangesReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.pendingChangesMutex.Lock()
	defer fake.pendingChangesMutex.Unlock()
	fake.PendingChangesStub = nil
	if fake.pendingChangesReturnsOnCall == nil {
		fake.pendingChangesReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.pendingChangesReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) ProductProperties(arg1 string) (io.Reader, error) {
	fake.productPropertiesMutex.Lock()
	ret, specificReturn := fake.productPropertiesReturnsOnCall[len(fake.productPropertiesArgsForCall)]
	fake.productPropertiesArgsForCall = append(fake.productPropertiesArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ProductPropertiesStub
	fakeReturns := fake.productPropertiesReturns
	fake.recordInvocation("ProductProperties", []interface{}{arg1})
	fake.productPropertiesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOmService) ProductPropertiesCallCount() int {
	fake.productPropertiesMutex.RLock()
	defer fake.productPropertiesMutex.RUnlock()
	return len(fake.productPropertiesArgsForCall)
}

func (fake *FakeOmService) ProductPropertiesCalls(stub func(string) (io.Reader, error)) {
	fake.productPropertiesMutex.Lock()
	defer fake.productPropertiesMutex.Unlock()
	fake.ProductPropertiesStub = stub
}

func (fake *FakeOmService) ProductPropertiesArgsForCall(i int) string {
	fake.productPropertiesMutex.RLock()
	defer fake.productPropertiesMutex.RUnlock()
	argsForCall := fake.productPropertiesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeOmService) ProductPropertiesReturns(result1 io.Reader, result2 error) {
	fake.productPropertiesMutex.Lock()
	defer fake.productPropertiesMutex.Unlock()
	fake.ProductPropertiesStub = nil
	fake.productPropertiesReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) ProductPropertiesReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.productPropertiesMutex.Lock()
	defer fake.productPropertiesMutex.Unlock()
	fake.ProductPropertiesStub = nil
	if fake.productPropertiesReturnsOnCall == nil {
		fake.productPropertiesReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.productPropertiesReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) ProductResources(arg1 string) (io.Reader, error) {
	fake.productResourcesMutex.Lock()
	ret, specificReturn := fake.productResourcesReturnsOnCall[len(fake.productResourcesArgsForCall)]
	fake.productResourcesArgsForCall = append(fake.productResourcesArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ProductResourcesStub
	fakeReturns := fake.productResourcesReturns
	fake.recordInvocation("ProductResources", []interface{}{arg1})
	fake.productResourcesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOmService) ProductResourcesCallCount() int {
	fake.productResourcesMutex.RLock()
	defer fake.productResourcesMutex.RUnlock()
	return len(fake.productResourcesArgsForCall)
}

func (fake *FakeOmService) ProductResourcesCalls(stub func(string) (io.Reader, error)) {
	fake.productResourcesMutex.Lock()
	defer fake.productResourcesMutex.Unlock()
	fake.ProductResourcesStub = stub
}

func (fake *FakeOmService) ProductResourcesArgsForCall(i int) string {
	fake.productResourcesMutex.RLock()
	defer fake.productResourcesMutex.RUnlock()
	argsForCall := fake.productResourcesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeOmService) ProductResourcesReturns(result1 io.Reader, result2 error) {
	fake.productResourcesMutex.Lock()
	defer fake.productResourcesMutex.Unlock()
	fake.ProductResourcesStub = nil
	fake.productResourcesReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) ProductResourcesReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.productResourcesMutex.Lock()
	defer fake.productResourcesMutex.Unlock()
	fake.ProductResourcesStub = nil
	if fake.productResourcesReturnsOnCall == nil {
		fake.productResourcesReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.productResourcesReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) VmTypes() (io.Reader, error) {
	fake.vmTypesMutex.Lock()
	ret, specificReturn := fake.vmTypesReturnsOnCall[len(fake.vmTypesArgsForCall)]
	fake.vmTypesArgsForCall = append(fake.vmTypesArgsForCall, struct {
	}{})
	stub := fake.VmTypesStub
	fakeReturns := fake.vmTypesReturns
	fake.recordInvocation("VmTypes", []interface{}{})
	fake.vmTypesMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOmService) VmTypesCallCount() int {
	fake.vmTypesMutex.RLock()
	defer fake.vmTypesMutex.RUnlock()
	return len(fake.vmTypesArgsForCall)
}

func (fake *FakeOmService) VmTypesCalls(stub func() (io.Reader, error)) {
	fake.vmTypesMutex.Lock()
	defer fake.vmTypesMutex.Unlock()
	fake.VmTypesStub = stub
}

func (fake *FakeOmService) VmTypesReturns(result1 io.Reader, result2 error) {
	fake.vmTypesMutex.Lock()
	defer fake.vmTypesMutex.Unlock()
	fake.VmTypesStub = nil
	fake.vmTypesReturns = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) VmTypesReturnsOnCall(i int, result1 io.Reader, result2 error) {
	fake.vmTypesMutex.Lock()
	defer fake.vmTypesMutex.Unlock()
	fake.VmTypesStub = nil
	if fake.vmTypesReturnsOnCall == nil {
		fake.vmTypesReturnsOnCall = make(map[int]struct {
			result1 io.Reader
			result2 error
		})
	}
	fake.vmTypesReturnsOnCall[i] = struct {
		result1 io.Reader
		result2 error
	}{result1, result2}
}

func (fake *FakeOmService) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.certificateAuthoritiesMutex.RLock()
	defer fake.certificateAuthoritiesMutex.RUnlock()
	fake.certificatesMutex.RLock()
	defer fake.certificatesMutex.RUnlock()
	fake.deployedProductsMutex.RLock()
	defer fake.deployedProductsMutex.RUnlock()
	fake.diagnosticReportMutex.RLock()
	defer fake.diagnosticReportMutex.RUnlock()
	fake.installationsMutex.RLock()
	defer fake.installationsMutex.RUnlock()
	fake.pendingChangesMutex.RLock()
	defer fake.pendingChangesMutex.RUnlock()
	fake.productPropertiesMutex.RLock()
	defer fake.productPropertiesMutex.RUnlock()
	fake.productResourcesMutex.RLock()
	defer fake.productResourcesMutex.RUnlock()
	fake.vmTypesMutex.RLock()
	defer fake.vmTypesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeOmService) recordInvocation(key string, args []interface{}) {
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

var _ opsmanager.OmService = new(FakeOmService)
