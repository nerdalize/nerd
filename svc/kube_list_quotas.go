package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	corev1 "k8s.io/api/core/v1"
)

//ListQuotaItem describes a namespace quota
type ListQuotaItem struct {
	RequestCPU    int64
	LimitCPU      int64
	LimitMemory   int64
	RequestMemory int64

	UseRequestCPU    int64
	UseLimitCPU      int64
	UseLimitMemory   int64
	UseRequestMemory int64
}

//NodeLimitedQuota is used when no quota is configured
var NodeLimitedQuota = ListQuotaItem{}

//ListQuotasInput is the input to ListQuotas
type ListQuotasInput struct{}

//ListQuotasOutput is the output to ListQuotas
type ListQuotasOutput struct {
	Items []*ListQuotaItem
}

//ListQuotas will list jobs on kubernetes
func (k *Kube) ListQuotas(ctx context.Context, in *ListQuotasInput) (out *ListQuotasOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	//Get the namespace quota
	quotas := &quotas{}
	err = k.visor.ListResources(ctx, kubevisor.ResourceTypeQuota, quotas, nil, nil)
	if err != nil {
		return nil, err
	}

	//get jobs and investivate
	out = &ListQuotasOutput{}
	for _, q := range quotas.Items {
		reqCPU, _ := q.Status.Hard[corev1.ResourceRequestsCPU]
		reqMem, _ := q.Status.Hard[corev1.ResourceRequestsMemory]
		limCPU, _ := q.Status.Hard[corev1.ResourceLimitsCPU]
		limMem, _ := q.Status.Hard[corev1.ResourceLimitsMemory]
		useReqCPU, _ := q.Status.Used[corev1.ResourceRequestsCPU]
		useReqMem, _ := q.Status.Used[corev1.ResourceRequestsMemory]
		useLimCPU, _ := q.Status.Used[corev1.ResourceLimitsCPU]
		useLimMem, _ := q.Status.Used[corev1.ResourceLimitsMemory]

		out.Items = append(out.Items, &ListQuotaItem{
			RequestCPU:    reqCPU.MilliValue(),
			RequestMemory: reqMem.MilliValue(),
			LimitCPU:      limCPU.MilliValue(),
			LimitMemory:   limMem.MilliValue(),

			UseLimitCPU:      useLimCPU.MilliValue(),
			UseRequestCPU:    useReqCPU.MilliValue(),
			UseLimitMemory:   useLimMem.MilliValue(),
			UseRequestMemory: useReqMem.MilliValue(),
		})
	}

	return out, nil
}

//quotas implements the list transformer interface to allow the kubevisor the manage names for us
type quotas struct{ *corev1.ResourceQuotaList }

func (quotas *quotas) Transform(fn func(in kubevisor.ManagedNames) (out kubevisor.ManagedNames)) {
	for i, j1 := range quotas.ResourceQuotaList.Items {
		quotas.Items[i] = *(fn(&j1).(*corev1.ResourceQuota))
	}
}

func (quotas *quotas) Len() int {
	return len(quotas.ResourceQuotaList.Items)
}
