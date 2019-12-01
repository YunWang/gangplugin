package predicate

import (
	"fmt"
	"github.com/YunWang/gangplugin/pkg/api/v1"
	v12 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type PodFilter struct {
	//predicate.Funcs
}

func (p *PodFilter) Create(createEvent event.CreateEvent) bool {

	if createEvent.Meta.GetNamespace()=="default" {
		if _,exist:=createEvent.Meta.GetAnnotations()[v1.GangKey];exist{
			return true
		}
	}
	return false
}
func (p *PodFilter) Update(updateEvent event.UpdateEvent) bool {
	if updateEvent.MetaNew.GetNamespace()=="default" {
		if _,exist:=updateEvent.MetaNew.GetAnnotations()[v1.GangKey];exist{
			oldPod:=updateEvent.MetaOld.(*v12.Pod)
			newPod:=updateEvent.MetaNew.(*v12.Pod)
			if oldPod.Status.Phase!=newPod.Status.Phase{
				fmt.Println("pod.Status.Phase changed")
				fmt.Println("OldPhase:"+oldPod.Status.Phase)
				fmt.Println("NewPhase:"+newPod.Status.Phase)
				return true
			}else if oldPod.DeletionTimestamp==nil && oldPod.DeletionTimestamp!=newPod.DeletionTimestamp{
				fmt.Println("DeleteTimestamp changed")
				fmt.Println("OldDeleteTime:")
				fmt.Println(oldPod.DeletionTimestamp)
				fmt.Println("NewDeleteTime:")
				fmt.Println(newPod.DeletionTimestamp)
				return true
			}
		}
	}
	return false
}
func (p *PodFilter) Delete(deleteEvent event.DeleteEvent) bool {
	//only status change should be passed and deletetimestamp changed should be passed
	//the former indicated pod's status changed
	//the latter indicated pod had been deleted
	fmt.Println("Delete pod")
	if deleteEvent.Meta.GetNamespace()=="default" {
		if _,exist:=deleteEvent.Meta.GetAnnotations()[v1.GangKey];exist{
			return true
		}
	}
	return false
}
func (p *PodFilter) Generic(genericEvent event.GenericEvent) bool {
	if genericEvent.Meta.GetNamespace()=="default" {
		if _,exist:=genericEvent.Meta.GetAnnotations()[v1.GangKey];exist{
			return true
		}
	}
	return false
}