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

package gang_controller

import (
	"context"
	"github.com/YunWang/gangplugin/pkg/predicate"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	batchv1 "github.com/YunWang/gangplugin/pkg/api/v1"
)

// GangReconciler reconciles a Gang object
type GangReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}



// +kubebuilder:rbac:groups=batch.wangyun.com,resources=gangs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch.wangyun.com,resources=gangs/status,verbs=get;update;patch

func (r *GangReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("gang", req.NamespacedName)

	gang := &batchv1.Gang{}
	err := r.Get(ctx, req.NamespacedName, gang)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if gang.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	podList:=v1.PodList{}
	if err:= r.List(ctx,&podList,client.InNamespace(gang.Namespace));err!=nil{
		return ctrl.Result{},err
	}
	succeedNum:=int32(0)
	runningNum:=int32(0)
	pendingNum:=int32(0)
	failedNum:=int32(0)
	unknownNum:=int32(0)
	for _,pod := range podList.Items {
		if gangName,exist:=pod.Annotations[batchv1.GangKey];exist && pod.DeletionTimestamp.IsZero() {
			if gangName == gang.Name {
				switch pod.Status.Phase {
				case v1.PodSucceeded:
					succeedNum+=1
				case v1.PodRunning:
					runningNum+=1
				case v1.PodPending:
					pendingNum+=1
				case v1.PodFailed:
					failedNum+=1
				case v1.PodUnknown:
					unknownNum+=1
				}
			}
		}
	}

	gang.Status.PodSucceeded=succeedNum
	gang.Status.PodFailed=failedNum
	gang.Status.PodRunning=runningNum
	gang.Status.PodPending=pendingNum
	gang.Status.PodUnknown=unknownNum

	minGang:=gang.Spec.MinGang

	if succeedNum==0&&pendingNum==0&&runningNum==0&&failedNum==0&&unknownNum==0{
		gang.Status.Phase=batchv1.GangPendingPhase
	}else{
		if succeedNum+runningNum<minGang{
			gang.Status.Phase=batchv1.GangUnknownPhase
		}else if succeedNum+runningNum>=minGang {
			if runningNum==0 && pendingNum==0 && unknownNum==0 && failedNum==0 && succeedNum>0{
				gang.Status.Phase=batchv1.GangCompletedPhase
			}else{
				gang.Status.Phase=batchv1.GangRunningPhase
			}
		}
	}

	oldGang:=&batchv1.Gang{}
	if err:=r.Get(ctx,types.NamespacedName{Namespace:gang.Namespace,Name:gang.Name},oldGang);err!=nil{
		return ctrl.Result{},err
	}
	if !reflect.DeepEqual(oldGang.Status,gang.Status){
		oldGang.Status=gang.Status
		if err:=r.Update(ctx,oldGang);err!=nil{
			return ctrl.Result{},err
		}
	}


	////deleteEvent
	//if gang.DeletionTimestamp!=nil {
	//	if index,ok:=utils.ContainsString(gang.Finalizers,batchv1.FinalizerKey);ok {
	//		//check PodTotal==0
	//		if gang.Status.PodTotal==0 {
	//			if err = r.deleteGangFinalizer(ctx,gang,index);err!=nil{
	//				return ctrl.Result{},err
	//			}
	//			return ctrl.Result{},nil
	//		}
	//		if err=r.removeDeleteTimestamp(ctx,gang);err!=nil{
	//			return ctrl.Result{},err
	//		}
	//		return ctrl.Result{},nil
	//	}
	//	return ctrl.Result{},nil
	//}
	//
	////createEvent
	//if gang.Status.Phase==""{
	//	if err=r.syncGangPhase(ctx,gang,batchv1.GangPendingPhase);err!=nil{
	//		return ctrl.Result{},err
	//	}
	//	return ctrl.Result{},nil
	//}
	//
	////updateEvent
	//minGang:=gang.Spec.MinGang
	//succeedNum:=gang.Status.PodSucceeded
	//runningNum:=gang.Status.PodRunning
	//pendingNum:=gang.Status.PodPending
	//failedNum:=gang.Status.PodFailed
	//unknownNum:=gang.Status.PodUnknown
	//if succeedNum+runningNum<minGang{
	//	if err=r.syncGangPhase(ctx,gang,batchv1.GangUnknownPhase);err!=nil{
	//		return ctrl.Result{},err
	//	}
	//	return ctrl.Result{},nil
	//}else if succeedNum+runningNum>=minGang {
	//	if runningNum==0 && pendingNum==0 && unknownNum==0 && failedNum==0 && succeedNum>0{
	//		if err=r.syncGangPhase(ctx,gang,batchv1.GangCompletedPhase);err!=nil {
	//			return ctrl.Result{},err
	//		}
	//		return ctrl.Result{},nil
	//	}else{
	//		if err=r.syncGangPhase(ctx,gang,batchv1.GangRunningPhase);err!=nil{
	//			return ctrl.Result{},err
	//		}
	//		return ctrl.Result{},nil
	//	}
	//}

	return ctrl.Result{}, nil
}

func (r *GangReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Gang{}).
		WithEventFilter(&predicate.GangFilter{}).
		Named("GangReconciler").
		Complete(r)
}

func NewGangReconcile(client client.Client,log logr.Logger,scheme *runtime.Scheme)*GangReconciler{
	return &GangReconciler{
		Client:client,
		Log:log,
		Scheme:scheme,
	}
}

func (r *GangReconciler)syncGangStatus(ctx context.Context,newGang *batchv1.Gang)error{
	oldGang := &batchv1.Gang{}
	err := r.Get(ctx, types.NamespacedName{Name: newGang.Name, Namespace: newGang.Namespace}, oldGang)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(oldGang.Status, newGang.Status) {
		oldGang.Status = newGang.Status
		if err = r.Update(ctx, oldGang); err != nil {
			return err
		}
	}
	return nil
}

func(r *GangReconciler)syncPodObjectMeta(ctx context.Context,pod *v1.Pod)error{
	oldPod:=&v1.Pod{}
	err:=r.Get(ctx,types.NamespacedName{Namespace:pod.Namespace,Name:pod.Name},oldPod)
	if err!=nil {
		return err
	}
	if !reflect.DeepEqual(oldPod.ObjectMeta,pod.ObjectMeta) {
		oldPod.ObjectMeta=pod.ObjectMeta
		if err=r.Update(ctx,oldPod);err!=nil {
			return err
		}
	}
	return nil
}



func (r *GangReconciler) deleteGangFinalizer(ctx context.Context,gang *batchv1.Gang,index int32)error{
	gang.Finalizers=append(gang.Finalizers[:index],gang.Finalizers[:index+1]...)

	oldGang:=&batchv1.Gang{}
	if err:=r.Get(ctx,types.NamespacedName{Namespace:gang.Namespace,Name:gang.Name},oldGang);err!=nil{
		return err
	}
	if !reflect.DeepEqual(oldGang.Finalizers,gang.Finalizers){
		oldGang.Finalizers=gang.Finalizers
		if err:=r.Update(ctx,oldGang);err!=nil{
			return err
		}
	}
	return nil
}

func (r *GangReconciler) removeDeleteTimestamp(ctx context.Context,gang *batchv1.Gang)error{
	gang.DeletionTimestamp=nil

	oldGang:=&batchv1.Gang{}
	if err:=r.Get(ctx,types.NamespacedName{Namespace:gang.Namespace,Name:gang.Name},oldGang);err!=nil{
		return err
	}
	if !reflect.DeepEqual(oldGang.DeletionTimestamp,gang.DeletionTimestamp){
		oldGang.DeletionTimestamp=gang.DeletionTimestamp
		if err:=r.Update(ctx,oldGang);err!=nil{
			return err
		}
	}
	return nil
}

func (r *GangReconciler) syncGangPhase(ctx context.Context,gang *batchv1.Gang,phase batchv1.GangPhase)error{
	gang.Status.Phase=phase
	if phase==batchv1.GangCompletedPhase {
		//set deleteTimestamp
		gang.DeletionTimestamp=&apimachineryv1.Time{
			Time:time.Now().Add(3*time.Minute),
		}
	}

	oldGang:=&batchv1.Gang{}
	if err:=r.Get(ctx,types.NamespacedName{Namespace:gang.Namespace,Name:gang.Name},oldGang);err!=nil{
		return err
	}
	if !reflect.DeepEqual(oldGang.Status.Phase,gang.Status.Phase){
		oldGang.Status.Phase=gang.Status.Phase
		if err:=r.Update(ctx,oldGang);err!=nil{
			return err
		}
	}
	return nil
}

func setGangDefault(gang *batchv1.Gang){
	gang.Status.PodRunning=0
	gang.Status.PodPending=0
	gang.Status.PodSucceeded=0
	gang.Status.PodUnknown=0
	gang.Status.PodFailed=0

}