package util

import (
	"fmt"
	"math"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1helper "k8s.io/kubernetes/pkg/apis/core/v1/helper"
	schedutil "k8s.io/kubernetes/pkg/scheduler/util"
)

// Resource is a collection of compute resource.
type Resource struct {
	MilliCPU         int64
	Memory           int64
	EphemeralStorage int64
	AllowedPodNumber int64

	// ScalarResources
	ScalarResources map[corev1.ResourceName]int64
}

// EmptyResource creates a empty resource object and returns.
func EmptyResource() *Resource {
	return &Resource{}
}

// NewResource creates a new resource object from resource list.
func NewResource(rl corev1.ResourceList) *Resource {
	r := &Resource{}
	for rName, rQuant := range rl {
		switch rName {
		case corev1.ResourceCPU:
			r.MilliCPU += rQuant.MilliValue()
		case corev1.ResourceMemory:
			r.Memory += rQuant.Value()
		case corev1.ResourcePods:
			r.AllowedPodNumber += rQuant.Value()
		case corev1.ResourceEphemeralStorage:
			r.EphemeralStorage += rQuant.Value()
		default:
			if schedutil.IsScalarResourceName(rName) {
				r.AddScalar(rName, rQuant.Value())
			}
		}
	}
	return r
}

// Add is used to add two resources.
func (r *Resource) Add(rl corev1.ResourceList) {
	if r == nil {
		return
	}

	for rName, rQuant := range rl {
		switch rName {
		case corev1.ResourceCPU:
			r.MilliCPU += rQuant.MilliValue()
		case corev1.ResourceMemory:
			r.Memory += rQuant.Value()
		case corev1.ResourcePods:
			r.AllowedPodNumber += rQuant.Value()
		case corev1.ResourceEphemeralStorage:
			r.EphemeralStorage += rQuant.Value()
		default:
			if schedutil.IsScalarResourceName(rName) {
				r.AddScalar(rName, rQuant.Value())
			}
		}
	}
}

// Sub is used to subtract two resources.
// Return error when the minuend is less than the subtrahend.
func (r *Resource) Sub(rl corev1.ResourceList) error {
	for rName, rQuant := range rl {
		switch rName {
		case corev1.ResourceCPU:
			cpu := rQuant.MilliValue()
			if r.MilliCPU < cpu {
				return fmt.Errorf("cpu difference is less than 0, remain %d, got %d", r.MilliCPU, cpu)
			}
			r.MilliCPU -= cpu
		case corev1.ResourceMemory:
			mem := rQuant.Value()
			if r.Memory < mem {
				return fmt.Errorf("memory difference is less than 0, remain %d, got %d", r.Memory, mem)
			}
			r.Memory -= mem
		case corev1.ResourcePods:
			pods := rQuant.Value()
			if r.AllowedPodNumber < pods {
				return fmt.Errorf("allowed pod difference is less than 0, remain %d, got %d", r.AllowedPodNumber, pods)
			}
			r.AllowedPodNumber -= pods
		case corev1.ResourceEphemeralStorage:
			ephemeralStorage := rQuant.Value()
			if r.EphemeralStorage < ephemeralStorage {
				return fmt.Errorf("allowed storage number difference is less than 0, remain %d, got %d", r.EphemeralStorage, ephemeralStorage)
			}
			r.EphemeralStorage -= ephemeralStorage
		default:
			if schedutil.IsScalarResourceName(rName) {
				rScalar, ok := r.ScalarResources[rName]
				scalar := rQuant.Value()
				if !ok {
					return fmt.Errorf("scalar resources %s does not exist, got %d", rName, scalar)
				}
				if rScalar < scalar {
					return fmt.Errorf("scalar resources %s difference is less than 0, remain %d, got %d", rName, rScalar, scalar)
				}
				r.ScalarResources[rName] = rScalar - scalar
			}
		}
	}
	return nil
}

// SetMaxResource compares with ResourceList and takes max value for each Resource.
func (r *Resource) SetMaxResource(rl corev1.ResourceList) {
	if r == nil {
		return
	}

	for rName, rQuant := range rl {
		switch rName {
		case corev1.ResourceCPU:
			if cpu := rQuant.MilliValue(); cpu > r.MilliCPU {
				r.MilliCPU = cpu
			}
		case corev1.ResourceMemory:
			if mem := rQuant.Value(); mem > r.Memory {
				r.Memory = mem
			}
		case corev1.ResourceEphemeralStorage:
			if ephemeralStorage := rQuant.Value(); ephemeralStorage > r.EphemeralStorage {
				r.EphemeralStorage = ephemeralStorage
			}
		case corev1.ResourcePods:
			if pods := rQuant.Value(); pods > r.AllowedPodNumber {
				r.AllowedPodNumber = pods
			}
		default:
			if schedutil.IsScalarResourceName(rName) {
				if value := rQuant.Value(); value > r.ScalarResources[rName] {
					r.SetScalar(rName, value)
				}
			}
		}
	}
}

// AddScalar adds a resource by a scalar value of this resource.
func (r *Resource) AddScalar(name corev1.ResourceName, quantity int64) {
	r.SetScalar(name, r.ScalarResources[name]+quantity)
}

// SetScalar sets a resource by a scalar value of this resource.
func (r *Resource) SetScalar(name corev1.ResourceName, quantity int64) {
	// Lazily allocate scalar resource map.
	if r.ScalarResources == nil {
		r.ScalarResources = map[corev1.ResourceName]int64{}
	}
	r.ScalarResources[name] = quantity
}

// ResourceList returns a resource list of this resource.
func (r *Resource) ResourceList() corev1.ResourceList {
	result := corev1.ResourceList{
		corev1.ResourceCPU:              *resource.NewMilliQuantity(r.MilliCPU, resource.DecimalSI),
		corev1.ResourceMemory:           *resource.NewQuantity(r.Memory, resource.BinarySI),
		corev1.ResourceEphemeralStorage: *resource.NewQuantity(r.EphemeralStorage, resource.BinarySI),
		corev1.ResourcePods:             *resource.NewQuantity(r.AllowedPodNumber, resource.DecimalSI),
	}
	for rName, rQuant := range r.ScalarResources {
		if v1helper.IsHugePageResourceName(rName) {
			result[rName] = *resource.NewQuantity(rQuant, resource.BinarySI)
		} else {
			result[rName] = *resource.NewQuantity(rQuant, resource.DecimalSI)
		}
	}
	return result
}

// MaxDivided returns how many replicas that the resource can be divided.
func (r *Resource) MaxDivided(rl corev1.ResourceList) int64 {
	res := int64(math.MaxInt64)
	for rName, rQuant := range rl {
		switch rName {
		case corev1.ResourceCPU:
			if cpu := rQuant.MilliValue(); cpu > 0 {
				res = MinInt64(res, r.MilliCPU/cpu)
			}
		case corev1.ResourceMemory:
			if mem := rQuant.Value(); mem > 0 {
				res = MinInt64(res, r.Memory/mem)
			}
		case corev1.ResourceEphemeralStorage:
			if ephemeralStorage := rQuant.Value(); ephemeralStorage > 0 {
				res = MinInt64(res, r.EphemeralStorage/ephemeralStorage)
			}
		default:
			if schedutil.IsScalarResourceName(rName) {
				rScalar, ok := r.ScalarResources[rName]
				if !ok {
					return 0
				}
				if scalar := rQuant.Value(); scalar > 0 {
					res = MinInt64(res, rScalar/scalar)
				}
			}
		}
	}
	res = MinInt64(res, r.AllowedPodNumber)
	return res
}

// LessEqual returns whether all dimensions of resources in r are less than or equal with that of rr.
func (r *Resource) LessEqual(rr *Resource) bool {
	lessEqualFunc := func(l, r int64) bool {
		return l <= r
	}

	if !lessEqualFunc(r.MilliCPU, rr.MilliCPU) {
		return false
	}
	if !lessEqualFunc(r.Memory, rr.Memory) {
		return false
	}
	if !lessEqualFunc(r.EphemeralStorage, rr.EphemeralStorage) {
		return false
	}
	if !lessEqualFunc(r.AllowedPodNumber, rr.AllowedPodNumber) {
		return false
	}
	for rrName, rrQuant := range rr.ScalarResources {
		rQuant := r.ScalarResources[rrName]
		if !lessEqualFunc(rQuant, rrQuant) {
			return false
		}
	}
	return true
}

// AddPodRequest add the effective request resource of a pod to the origin resource.
// The Pod's effective request is the higher of:
// - the sum of all app containers(spec.Containers) request for a resource.
// - the effective init containers(spec.InitContainers) request for a resource.
// The effective init containers request is the highest request on all init containers.
func (r *Resource) AddPodRequest(podSpec *corev1.PodSpec) *Resource {
	for _, container := range podSpec.Containers {
		r.Add(container.Resources.Requests)
	}
	for _, container := range podSpec.InitContainers {
		r.SetMaxResource(container.Resources.Requests)
	}
	return r
}

// AddResourcePods adds pod resources into the Resource.
// Notice that a pod request resource list does not contain a request for pod resources,
// this function helps to add the pod resources.
func (r *Resource) AddResourcePods(pods int64) {
	r.Add(corev1.ResourceList{
		corev1.ResourcePods: *resource.NewQuantity(pods, resource.DecimalSI),
	})
}

// MinInt64 returns the smaller of two int64 numbers.
func MinInt64(a, b int64) int64 {
	if a <= b {
		return a
	}
	return b
}
