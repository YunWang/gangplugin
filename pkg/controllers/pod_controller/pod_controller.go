package pod_controller

import (
	"context"
	batchv1 "github.com/YunWang/gangplugin/pkg/api/v1"
	"github.com/YunWang/gangplugin/pkg/predicate"
	"github.com/YunWang/gangplugin/pkg/utils"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PodReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	lastSeenPhase map[types.NamespacedName]corev1.PodPhase //pod.NamespacedName->gang.Name
}

func (pr *PodReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := pr.Log.WithValues("pod", req.NamespacedName)

	//2.
	//3.
	//4.if pod is running, then running pluses 1
	//5.if pod completes, then succeeded pluses + 1 and running subtracts 1
	//6.validate gang status,total=running+pending+succeeded+failed+unknown
	//7.sync gang's status



	//1.get pod
	pod := &corev1.Pod{}
	if err := pr.Get(ctx, req.NamespacedName, pod);err!=nil{
		if errors.IsNotFound(err){
			log.V(1).Info("Pod was deleted")
			return ctrl.Result{},nil
		}
		return ctrl.Result{}, err
	}


	//2.get GangName
	gangName,exist:=pod.ObjectMeta.Annotations[batchv1.GangKey]
	if !exist {
		return ctrl.Result{},nil
	}
	//3.get gang
	gang:=&batchv1.Gang{}
	if err:=pr.Get(ctx,types.NamespacedName{Name:gangName,Namespace:pod.Namespace},gang);err!=nil{
		return ctrl.Result{},err
	}
	//4.get lastPhase
	lastPhase:=pr.lastSeenPhase[req.NamespacedName]

	//wo only care aboud two events. First, deleteTimestamp changed.Second
	//pod.status.phase changed
	//5.deleteTimestamp changed, that means pod would be deleted
	if !pod.DeletionTimestamp.IsZero() {
		//deletePodEvent
		if index,ok:=utils.ContainsString(pod.ObjectMeta.Finalizers,batchv1.FinalizerKey);ok{
			//update gang status
			if err:=pr.syncGangStatus(ctx,gang,lastPhase,"");err!=nil{
				return ctrl.Result{},err
			}
			//delete pod.NamespacedName from lastSeenPhase
			delete(pr.lastSeenPhase,req.NamespacedName)
			//delete finalizer from pod.Finalizers
			if err:=pr.removePodFinalizer(ctx,pod,index);err!=nil{
				return ctrl.Result{},err
			}
		}else{

		}
		return ctrl.Result{},nil
	}
	//6.pod status.phase changed,that means pod's phase changed
	if err:=pr.syncGangStatus(ctx,gang,lastPhase,pod.Status.Phase);err!=nil{
		return ctrl.Result{},err
	}
	//7.update lastSeenPhase
	pr.lastSeenPhase[req.NamespacedName]=pod.Status.Phase

	return ctrl.Result{}, nil
}

func (pr *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(&predicate.PodFilter{}).
		Complete(pr)
}

func NewPodController(client client.Client, log logr.Logger, scheme *runtime.Scheme) *PodReconciler {
	return &PodReconciler{
		Client:        client,
		Log:           log,
		Scheme:        scheme,
		lastSeenPhase: make(map[types.NamespacedName]corev1.PodPhase),
	}
}


func (pr *PodReconciler) syncGangStatus(ctx context.Context, gang *batchv1.Gang,lastPhase,currentPhase corev1.PodPhase) error {
	pr.calculateGangStatus(gang,lastPhase,currentPhase)

	oldGang := &batchv1.Gang{}
	err := pr.Get(ctx, types.NamespacedName{Name: gang.Name, Namespace: gang.Namespace}, oldGang)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(oldGang.Status, gang.Status) {
		oldGang.Status = gang.Status
		if err = pr.Update(ctx, oldGang); err != nil {
			pr.Log.Info("Update Gang failed")
			return err
		}
	}
	return nil
}

func (pr *PodReconciler) removePodFinalizer(ctx context.Context,pod *corev1.Pod,finalizerIndex int32)error{
	pod.Finalizers=append(pod.Finalizers[:finalizerIndex],pod.Finalizers[finalizerIndex+1:]...)

	oldPod:=&corev1.Pod{}
	if err:=pr.Get(ctx,types.NamespacedName{Namespace:pod.Namespace,Name:pod.Name},oldPod);err!=nil{
		return err
	}
	if !reflect.DeepEqual(oldPod.Finalizers,pod.Finalizers){
		oldPod.Finalizers=pod.Finalizers
		if err:=pr.Update(ctx,oldPod);err!=nil{
			return err
		}
	}
	return nil
}

func (pr *PodReconciler)calculateGangStatus(gang *batchv1.Gang,lastPhase,currentPhase corev1.PodPhase){
	if lastPhase ==currentPhase{
		return
	}
	//update lastPhase
	if lastPhase!=""{
		pr.calculatePodNumber(gang,lastPhase,-1)
	}
	//update currentPhase
	if currentPhase!=""{
		pr.calculatePodNumber(gang,currentPhase,1)
	}
}

func (pr *PodReconciler)calculatePodNumber(gang *batchv1.Gang,phase corev1.PodPhase,offset int32){
	switch phase {
	case corev1.PodRunning:
		gang.Status.PodRunning+=offset
	case corev1.PodPending:
		gang.Status.PodPending+=offset
	case corev1.PodFailed:
		gang.Status.PodFailed+=offset
	case corev1.PodUnknown:
		gang.Status.PodUnknown+=offset
	case corev1.PodSucceeded:
		gang.Status.PodSucceeded+=offset
	}
}
