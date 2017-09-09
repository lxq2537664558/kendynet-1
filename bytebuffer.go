package kendynet

import (
	"fmt"
)

func IsPow2(size uint64) bool{
	return (size&(size-1)) == 0
}

func SizeofPow2(size uint64) uint64{
	if IsPow2(size){
		return size
	}
	size = size -1
	size = size-1
	size = size | (size>>1)
	size = size | (size>>2)
	size = size | (size>>4)
	size = size | (size>>8)
	size = size | (size>>16)
	return size + 1
}

func GetPow2(size uint64) uint64{
	var pow2 uint64 = 0
	if !IsPow2(size) {
		size = (size << 1)
	}
	for size > 1 {
		pow2++
	}
	return pow2
}

const (
	MaxBuffSize uint64 = 0xFFFFFFFF 
)

type ByteBuffer struct {
	buffer []byte
	datasize uint64
	capacity uint64
	needcopy bool    //标记是否执行写时拷贝
}

var (
	ErrMaxSizeExceeded     = fmt.Errorf("bytebuffer: Max Buffer Size Exceeded")
	ErrInvaildAgr          = fmt.Errorf("bytebuffer: Invaild Idx or size")
)

func NewByteBuffer(arg ...interface{})(*ByteBuffer){
	if len(arg) < 2 {
		var size uint64
		if len(arg) == 0 {
			size = 128
		} else {
			switch arg[0].(type) {
				case int8: 
					size = (uint64)(arg[0].(int8))
					break
				case uint8: 
					size = (uint64)(arg[0].(uint8))
					break
				case int16: 
					size = (uint64)(arg[0].(int16))
					break
				case uint16: 
					size = (uint64)(arg[0].(uint16))
					break
				case int32: 
					size = (uint64)(arg[0].(int32))
					break
				case uint32: 
					size = (uint64)(arg[0].(uint32))
					break
				case int64: 
					size = (uint64)(arg[0].(int64))
					break
				case uint64: 
					size = (uint64)(arg[0].(uint64))
					break
				default:
					return nil
			}
		}
		return &ByteBuffer{buffer:make([]byte,size),datasize:0,capacity:size,needcopy:false}
	} else if len(arg) == 2 {
		var bytes []byte
		var size uint64
		switch arg[0].(type) {
			case []byte:
				bytes = arg[0].([]byte)
				break
			default:
				return nil
		}
		switch arg[1].(type) {
			case int8: 
				size = (uint64)(arg[1].(int8))
				break
			case uint8: 
				size = (uint64)(arg[1].(uint8))
				break
			case int16: 
				size = (uint64)(arg[1].(int16))
				break
			case uint16: 
				size = (uint64)(arg[1].(uint16))
				break
			case int32: 
				size = (uint64)(arg[1].(int32))
				break
			case uint32: 
				size = (uint64)(arg[1].(uint32))
				break
			case int64: 
				size = (uint64)(arg[1].(int64))
				break
			case uint64: 
				size = (uint64)(arg[1].(uint64))
				break
			default:
				return nil
		}
		/*
		 * 直接引用bytes,并设置needcopy标记
		 * 如果ByteBuffer要修改bytes中的内容，首先要先执行拷贝，之后才能修改
		*/
		return &ByteBuffer{buffer:bytes,datasize:size,capacity:(uint64)(cap(bytes)),needcopy:true}
	} else {
		return nil
	}
}

func (this *ByteBuffer) Reset() {
	this.datasize = 0
}

func (this *ByteBuffer) Clone() (*ByteBuffer){
	b := make([]byte,this.capacity)
	copy(b[0:],this.buffer[:this.capacity])
	return &ByteBuffer{buffer:b,datasize:this.datasize,capacity:this.capacity,needcopy:false}
}

func (this *ByteBuffer) Buffer()([]byte){
	return this.buffer
}

func (this *ByteBuffer) Len()(uint64){
	return this.datasize
}

func (this *ByteBuffer) Cap()(uint64){
	return this.capacity
}

func (this *ByteBuffer) expand(newsize uint64)(error){
	newsize = SizeofPow2(newsize)
	if newsize > MaxBuffSize {
		return ErrMaxSizeExceeded
	}
	//allocate new buffer
	tmpbuf := make([]byte,newsize)
	//copy data
	copy(tmpbuf[0:], this.buffer[:this.datasize])
	//replace buffer
	this.buffer = tmpbuf
	this.capacity = newsize
	return nil
}

func (this *ByteBuffer) checkCapacity(idx,size uint64)(error){
	if size >= this.capacity && idx + size < this.capacity {
		//溢出
		return ErrMaxSizeExceeded
	}

	if this.needcopy {
		//需要执行写时拷贝
		sizeneed := idx + size
		if sizeneed > MaxBuffSize {
			return ErrMaxSizeExceeded
		}
		//allocate new buffer
		tmpbuf := make([]byte,sizeneed)
		//copy data
		copy(tmpbuf[0:], this.buffer[:this.datasize])
		//replace buffer
		this.buffer = tmpbuf
		this.capacity = sizeneed
		this.needcopy = false
		return nil
	}

	if idx + size > this.capacity {
		err := this.expand(idx+size)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *ByteBuffer) PutBytes(idx uint64,value []byte)(error){
	sizeneed := (uint64)(len(value))
	err := this.checkCapacity(idx,sizeneed)
	if err != nil {
		return err
	}
	copy(this.buffer[idx:],value[:sizeneed])
	if idx + sizeneed > this.datasize {
		this.datasize = idx + sizeneed
	}
	return nil
}

func (this *ByteBuffer) GetBytes(idx uint64,size uint64) (ret []byte,err error) {
	ret = nil
	err = nil
	if size >= this.datasize && idx + size < this.datasize {
		err = ErrInvaildAgr
		return
	}
	if idx + size > this.datasize {
		err = ErrInvaildAgr
		return
	}
	ret = this.buffer[idx:idx + size]
	return
}

func (this *ByteBuffer) PutString(idx uint64,value string)(error){
	sizeneed := (uint64)(len(value))
	err := this.checkCapacity(idx,sizeneed)
	if err != nil {
		return err
	}
	copy(this.buffer[idx:],value[:sizeneed])
	if idx + sizeneed > this.datasize {
		this.datasize = idx + sizeneed
	}
	return nil
}

func (this *ByteBuffer) GetString(idx uint64,size uint64) (ret string,err error) {
	var bytes []byte
	bytes,err = this.GetBytes(idx,size)
	if bytes != nil {
		ret = string(bytes)
	}
	return
}

/*
func main() {
	buff := NewByteBuffer()
	buff.PutString(0,"hello")
	ret,err := buff.GetString(0,5)
	if nil == err {
		fmt.Printf("%s\n",ret)
	}
}
*/


