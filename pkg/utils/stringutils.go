package utils

func ContainsString(strs []string,key string)(int32,bool){
	for index,elem:=range strs{
		if elem==key {
			return int32(index),true
		}
	}
	return -1,false
}