package controllers

import (
	"context"
	gang "github.com/YunWang/gangplugin/pkg/api/v1"
	logModule "github.com/YunWang/gangplugin/pkg/log"
	"github.com/YunWang/gangplugin/pkg/predicate"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

type PodReconciler struct {
	client.Client
	RWLock        sync.RWMutex
	Log           logr.Logger
	Scheme        *runtime.Scheme
	lastSeenPhase map[types.NamespacedName]corev1.PodPhase
	belongTo      map[types.NamespacedName]string
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
	err := pr.Get(ctx, req.NamespacedName, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			//3.pod has been deleted,then total subtracts 1 and pending subtracts 1
			gangName, exist := pr.belongTo[req.NamespacedName]
			if !exist {
				return ctrl.Result{}, nil
			}
			newGang := &gang.Gang{}
			err = pr.Get(ctx, types.NamespacedName{Name: gangName, Namespace: req.Namespace}, newGang)
			if err != nil {
				delete(pr.belongTo, req.NamespacedName)
				if _, exist := pr.lastSeenPhase[req.NamespacedName]; exist {
					delete(pr.lastSeenPhase, req.NamespacedName)
				}
				return ctrl.Result{}, nil
			}
			//get last phase
			lastPhase := pr.lastSeenPhase[req.NamespacedName]
			//update gang's status
			newGang.Status.Total -= 1
			update(newGang, lastPhase, -1)

			if !newGang.Validate() {
				log.V(logModule.Trace).Info("Failed to validate gang's status,total!=running+pending+succeeded+failed+unknow!")
				return ctrl.Result{}, errors.NewBadRequest("total!=running+pending+succeeded+failed+unknow")
			}

			err = pr.syncGangStatus(ctx, newGang)
			if err != nil {
				log.V(logModule.Trace).Info("Failed to sync Gang status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
		log.V(logModule.Trace).Info("Failed to get pod")
		return ctrl.Result{}, err
	}

	//2.check which gang this pod belong to, and update gang's status
	name, exist := pod.Annotations[gang.GangKey]
	if !exist {
		log.V(logModule.Debug).Info("Pod isn't managed by any gang!")
		return ctrl.Result{}, nil
	}

	//log.V(logModule.Trace).Info("Begin to reconcile pod and update gang's status")
	newGang := &gang.Gang{}
	err2 := pr.Get(ctx, types.NamespacedName{Name: name, Namespace: pod.Namespace}, newGang)
	if err2 != nil {
		log.V(logModule.Trace).Info("Failed to get gang which the pod with annotation{value:" + name + "} belongs to")
		log.V(logModule.Trace).Info("It seems gang had been deleted")
		return ctrl.Result{}, err2
	}

	//4&5.
	if lastPodPhase, exist := pr.lastSeenPhase[types.NamespacedName{Namespace: pod.Namespace, Name: pod.Name}]; !exist {
		//new pod is been created
		newGang.Status.Total += 1
		update(newGang, pod.Status.Phase, 1)
		pr.lastSeenPhase[types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}] = pod.Status.Phase
		pr.belongTo[req.NamespacedName] = newGang.Name
	} else {
		updateGangStatus(newGang, pod.Status.Phase, lastPodPhase)
		pr.lastSeenPhase[types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}] = pod.Status.Phase
	}
	//6.
	if !newGang.Validate() {
		log.V(logModule.Trace).Info("Failed to validate gang's status,total!=running+pending+succeeded+failed+unknow!")
		return ctrl.Result{}, errors.NewBadRequest("total!=running+pending+succeeded+failed+unknow")
	}
	//7.
	err = pr.syncGangStatus(ctx, newGang)
	if err != nil {
		log.V(logModule.Trace).Info("Failed to sync Gang status")
		return ctrl.Result{}, err
	}

	//log.V(logModule.Trace).Info("Reconcile gang successfully!")

	return ctrl.Result{}, nil
}

func (pr *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(&predicate.NamespacePredicate{}).
		Complete(pr)
}

func NewPodController(client client.Client, log logr.Logger, scheme *runtime.Scheme) *PodReconciler {
	return &PodReconciler{
		Client:        client,
		Log:           log,
		Scheme:        scheme,
		lastSeenPhase: make(map[types.NamespacedName]corev1.PodPhase),
		belongTo:      make(map[types.NamespacedName]string),
	}
}

func updateGangStatus(gang *gang.Gang, currentPhase, lastPhase corev1.PodPhase) {
	if currentPhase != lastPhase {
		update(gang, currentPhase, 1)
		update(gang, lastPhase, -1)
	}
}

func update(gang *gang.Gang, phase corev1.PodPhase, operation int32) {
	switch phase {
	case corev1.PodRunning:
		gang.Status.Running += operation
	case corev1.PodSucceeded:
		gang.Status.Succeeded += operation
	case corev1.PodPending:
		gang.Status.Pending += operation
	case corev1.PodFailed:
		gang.Status.Failed += operation
	case corev1.PodUnknown:
		gang.Status.Unknown += operation
	}
}

func (pr *PodReconciler) syncGangStatus(ctx context.Context, newGang *gang.Gang) error {
	oldGang := &gang.Gang{}
	err := pr.Get(ctx, types.NamespacedName{Name: newGang.Name, Namespace: newGang.Namespace}, oldGang)
	if err != nil {
		return err
	}
	//log.V(logModule.Trace).Info("old=?=new", "result:", reflect.DeepEqual(oldGang, newGang))
	if !reflect.DeepEqual(oldGang.Status, newGang.Status) {
		oldGang.Status = newGang.Status
		pr.RWLock.Lock()
		defer pr.RWLock.Unlock()
		if err = pr.Update(ctx, oldGang); err != nil {
			return err
		}
	}
	return nil
}

//default is 0 for all status
func setDefaultStatus(g *gang.Gang) error {
	g.Status.Total = 0
	g.Status.Running = 0
	g.Status.Succeeded = 0
	g.Status.Pending = 0
	g.Status.Failed = 0
	g.Status.Unknown = 0
	if !g.Validate() {
		return errors.NewBadRequest("Total!=running+Succeeded+pending+failed+unknown")
	}
	return nil
}
