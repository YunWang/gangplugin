/*
Copyright 2019 wangyun.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	logModule "github.com/YunWang/gangplugin/pkg/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sync"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "github.com/YunWang/gangplugin/pkg/api/v1"
)

// GangReconciler reconciles a Gang object
type GangReconciler struct {
	client.Client
	RWLock sync.RWMutex
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=batch.wangyun.com,resources=gangs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.wangyun.com,resources=gangs/status,verbs=get;update;patch

func (r *GangReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("gang", req.NamespacedName)

	// your logic here
	//1.get gang
	//2.set default gang status
	//3.sync status
	log.V(logModule.Trace).Info("Begin to reconcile Gang{Namespace:" + req.Namespace + ",Name:" + req.Name + "}")

	//1.
	gang := &batchv1.Gang{}
	err := r.Get(ctx, req.NamespacedName, gang)
	if err != nil {
		if errors.IsNotFound(err) {
			log.V(logModule.Trace).Info("Gang has been deleted!")
			return ctrl.Result{}, nil
		}
		log.V(logModule.Trace).Info("Failed to get Gang{Namespace:" + req.Namespace + ",Name:" + req.Name + "}")
		return ctrl.Result{}, err
	}

	//2.
	err = r.setDefaultStatus(gang)
	if err != nil {
		return ctrl.Result{}, err
	}

	//3.
	oldGang := &batchv1.Gang{}
	err = r.Get(ctx, types.NamespacedName{Namespace: gang.Namespace, Name: gang.Name}, oldGang)
	if err != nil {
		log.V(logModule.Trace).Info("Failed to get old gang")
		return ctrl.Result{}, err
	}

	if !reflect.DeepEqual(oldGang.Status, gang.Status) {
		oldGang.Status = gang.Status
		r.RWLock.Lock()
		defer r.RWLock.Unlock()
		if err := r.Status().Update(ctx, oldGang); err != nil {
			log.V(logModule.Trace).Info("Failed to update gang")
			return ctrl.Result{}, err
		}
	}

	log.V(logModule.Trace).Info("Reconcile gang successfully!")

	return ctrl.Result{}, nil
}

func (r *GangReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Gang{}).
		Complete(r)
}

//default is 0 for all status
func (r *GangReconciler) setDefaultStatus(g *batchv1.Gang) error {
	g.Status.Total = 0
	g.Status.Running = 0
	g.Status.Succeeded = 0
	g.Status.Pending = 0
	g.Status.Failed = 0
	g.Status.Unknown = 0
	if !g.Validate() {
		r.Log.V(logModule.Trace).Info("Failed to set default status for gang,because total!=running+succeeded")
		return errors.NewBadRequest("Total!=running+Succeeded+pending+failed+unknown")
	}
	return nil
}
