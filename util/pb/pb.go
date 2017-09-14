package pb

import (
	"github.com/golang/protobuf/proto"
	"github.com/sniperHW/kendynet"
	"fmt"
	"reflect"
	"sync"
)

const (
	PBHeaderSize uint64 = 4
	PBIdSize uint64 = 4
)

var id uint32 = 1
var nameToTypeID map[string]uint32 = make(map[string]uint32)
var idToMeta map[uint32]reflect.Type = make(map[uint32]reflect.Type)
var mutex *sync.Mutex = new(sync.Mutex)

func newMessage(id uint32) (msg proto.Message,err error){
    defer func(){
   		mutex.Unlock()
    }()
    mutex.Lock()
   if mt,ok := idToMeta[id];ok{
          msg = reflect.New(mt.Elem()).Interface().(proto.Message)
   } else{
          err = fmt.Errorf("not found %d",id)
   }
   return
}

//根据名字注册实例
func Register(msg proto.Message) (err error) {
    defer func(){
   		mutex.Unlock()
    }()
    mutex.Lock()
	tt := reflect.TypeOf(msg)
	name := tt.String()

	if _,ok := nameToTypeID[name];ok {
		err = fmt.Errorf("%s already register",name)
		return
	}

    nameToTypeID[name] = id
    idToMeta[id] = tt
    id++
    return nil
}


func Encode(o interface{},maxMsgSize uint64) (r *kendynet.ByteBuffer,e error) {
    mutex.Lock()
	typeID,ok := nameToTypeID[reflect.TypeOf(o).String()]
	mutex.Unlock()
	if !ok {
		e = fmt.Errorf("unregister type:%s",reflect.TypeOf(o).String())
	}

	msg := o.(proto.Message)

	data, err := proto.Marshal(msg)
	if err != nil {
		e = err
		return
	}

	dataLen := uint64(len(data))
	if dataLen  > maxMsgSize {
		e = fmt.Errorf("message size limite maxMsgSize[%d],msg payload[%d]",maxMsgSize,dataLen)
		return
	}

	totalLen := PBHeaderSize + PBIdSize + dataLen

	buff := kendynet.NewByteBuffer(totalLen)
	//写payload大小
	buff.AppendUint32(uint32(totalLen - PBHeaderSize))
	//写类型ID
	buff.AppendUint32(typeID)
	//写数据
	buff.AppendBytes(data)
	r = buff
	return
}

func Decode(buff []byte,start uint64,end uint64,maxMsgSize uint64) (proto.Message,uint64,error) {

	dataLen := end - start

	if dataLen < PBHeaderSize {
		return nil,0,nil
	}

	reader := kendynet.NewByteBuffer(buff[start:end],dataLen)

	s := uint64(0)


	payload,err := reader.GetUint32(0)

	if err != nil {
		return nil,0,err
	}

	if uint64(payload) > maxMsgSize {
		return nil,0,fmt.Errorf("Decode size limited maxMsgSize[%d],msg payload[%d]",maxMsgSize,payload)
	}else if uint64(payload) == 0 {
		return nil,0,fmt.Errorf("Decode header payload == 0")
	}

	totalPacketSize := uint64(payload) + PBHeaderSize

	if totalPacketSize > dataLen {
		return nil,0,nil
	}

	s += PBHeaderSize	

	typeID,_ := reader.GetUint32(s)

	msg,err := newMessage(typeID) 

	if err != nil {
		return nil,0,fmt.Errorf("unregister type:%d",typeID)
	}

	s += PBIdSize

	pbDataLen := totalPacketSize - PBHeaderSize - PBIdSize

	pbData,_ := reader.GetBytes(s,pbDataLen)

	err = proto.Unmarshal(pbData, msg)

	if err != nil {
		return nil,0,err
	}

	return msg,totalPacketSize,nil

} 