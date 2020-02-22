// MIT License
//
// Copyright (c) Microsoft Corporation. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE

package api

import (
	"fmt"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

///////////////////////////////////////////////////////////////////////////////////////
// General Types
///////////////////////////////////////////////////////////////////////////////////////
type (
	CellType      string
	CellAddress   string
	ReservationId string
)

// Physical cluster definition
type PhysicalClusterSpec struct {
	CellTypes     map[CellType]CellTypeSpec `yaml:"cellTypes"`
	PhysicalCells []PhysicalCellSpec        `yaml:"physicalCells"`
}

type CellTypeSpec struct {
	ChildCellType   CellType `yaml:"childCellType"`
	ChildCellNumber int32    `yaml:"childCellNumber"`
	IsNodeLevel     bool     `yaml:"isNodeLevel"`
}

// Specify physical Cell instances.
type PhysicalCellSpec struct {
	CellType      CellType           `yaml:"cellType"`
	CellAddress   CellAddress        `yaml:"cellAddress"`
	ReservationId ReservationId      `yaml:"reservationId"`
	CellChildren  []PhysicalCellSpec `yaml:"cellChildren,omitempty"`
}

// Virtual cluster definition
type VirtualClusterName string

type VirtualClusterSpec struct {
	VirtualCells  []VirtualCellSpec  `yaml:"virtualCells"`
	ReservedCells []ReservedCellSpec `yaml:"reservedCells,omitempty"`
}

type VirtualCellSpec struct {
	CellNumber int32    `yaml:"cellNumber"`
	CellType   CellType `yaml:"cellType"`
}

type ReservedCellSpec struct {
	ReservationId ReservationId `yaml:"reservationId"`
}

type PodSchedulingSpec struct {
	VirtualCluster       VirtualClusterName `yaml:"virtualCluster"`
	Priority             int32              `yaml:"priority"`
	ReservationId        ReservationId      `yaml:"reservationId"`
	GpuType              string             `yaml:"gpuType"`
	GpuNumber            int32              `yaml:"gpuNumber"`
	GangReleaseEnable    bool               `yaml:"gangReleaseEnable"`
	LazyPreemptionEnable bool               `yaml:"lazyPreemptionEnable"`
	AffinityGroup        *AffinityGroupSpec `yaml:"affinityGroup"`
}

type AffinityGroupSpec struct {
	Name    string                    `yaml:"name"`
	Members []AffinityGroupMemberSpec `yaml:"members"`
}

type AffinityGroupMemberSpec struct {
	PodNumber int32 `yaml:"podNumber"`
	GpuNumber int32 `yaml:"gpuNumber"`
}

// Used to recover scheduler allocated resource
type PodBindInfo struct {
	Node                  string                        `yaml:"node"`         // node to bind
	GpuIsolation          []int32                       `yaml:"gpuIsolation"` // GPUs to bind
	CellChain             string                        `yaml:"cellChain"`    // cell chain selected
	AffinityGroupBindInfo []AffinityGroupMemberBindInfo `yaml:"affinityGroupBindInfo"`
}

type AffinityGroupMemberBindInfo struct {
	PodPlacements []PodPlacementInfo `yaml:"podPlacements"`
}

type PodPlacementInfo struct {
	PhysicalNode       string  `yaml:"physicalNode"`
	PhysicalGpuIndices []int32 `yaml:"physicalGpuIndices"`
	// preassigned cell types used by the pods. used to locate the virtual cells
	// when adding an allocated pod
	PreassignedCellTypes []CellType `yaml:"preassignedCellTypes"`
}

type WebServerPaths struct {
	Paths []string `json:"paths"`
}

type WebServerError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewWebServerError(code int, message string) *WebServerError {
	return &WebServerError{
		Code:    code,
		Message: message,
	}
}

func (err *WebServerError) Error() string {
	return fmt.Sprintf("Code: %v, Message: %v", err.Code, err.Message)
}

// WebServer Exposed Objects: Align with K8S Objects
type ObjectMeta struct {
	Name string `json:"name"`
}

type AffinityGroupList struct {
	Items []AffinityGroup `json:"items"`
}

type AffinityGroup struct {
	ObjectMeta `json:"metadata"`
	Status     AffinityGroupStatus `json:"status"`
}

type AffinityGroupStatus struct {
	LazyPreemptionStatus *LazyPreemptionStatus `json:"lazyPreemptionStatus"`
}

type LazyPreemptionStatus struct {
	// The AffinityGroup who has lazy preempted it.
	Preemptor string `json:"preemptor"`
	// It was lazy preempted at PreemptionTime.
	PreemptionTime meta.Time `json:"preemptionTime"`
}

type ClusterStatus struct {
	// Status of cells in the physical cluster
	PhysicalCluster []*PhysicalCellStatus `json:"physicalCluster"`
	// Status of cells in each VC
	VirtualClusters map[string][]*VirtualCellStatus `json:"virtualClusters"`
}

type CellStatus struct {
	GpuType  string `json:"gpuType,omitempty"`
	CellType string `json:"cellType"`
	// Address of a physical cell consists of its address (or index) in each level
	// (e.g., node0/0/0/0 may represent node0, CPU socket 0, PCIe switch 0, GPU 0.
	// Address of a virtual cell consists of its VC name, index of the preassigned cell,
	// and the relative index in each level inside the preassigned cell
	// (e.g., VC1/0/0 may represent VC1, preassigned cell 0, index 0 among its children)
	CellAddress string `json:"cellAddress"`
	// State can be one of "Free", "Used", and "Bad"
	State    string `json:"state"`
	Priority int32  `json:"priority"`
}

type PhysicalCellStatus struct {
	CellStatus
	Children    []*PhysicalCellStatus `json:"children,omitempty"`
	Vc          string                `json:"vc,omitempty"`
	VirtualCell *VirtualCellStatus    `json:"virtualCell,omitempty"`
}

type VirtualCellStatus struct {
	CellStatus
	Children     []*VirtualCellStatus `json:"children,omitempty"`
	PhysicalCell *PhysicalCellStatus  `json:"physicalCell,omitempty"`
}

func GenerateOpporVirtualCell(pc *PhysicalCellStatus) *VirtualCellStatus {
	vc := &VirtualCellStatus{
		CellStatus: CellStatus{
			GpuType:     pc.GpuType,
			CellType:    pc.CellType,
			CellAddress: pc.CellAddress + "-opp",
			State:       UsedState,
			Priority:    OpportunisticPriority,
		},
		PhysicalCell: pc,
	}
	return vc
}
