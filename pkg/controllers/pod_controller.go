package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type PodReconcile struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (pr *PodReconcile) Reconcile(req ctrl.Request) (ctrl.Result, error){
	ctx := context.Background()
	log := pr.Log.WithValues("pod", req.NamespacedName)

	//1.get pod
	//2.check status
	//

	return ctrl.Result{},nil
}

func (pr *PodReconcile) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(pr)
}