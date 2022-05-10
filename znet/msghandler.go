package znet

import (
	"fmt"
	"strconv"
	"zinx/utils"
	"zinx/ziface"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter // 存放每个 msgID 对应的 handle
	WorkerPoolSize uint32                    // 工作池 (协程池)
	TaskQueue      []chan ziface.IRequset    // 任务队列 (一个 worker 对应一个)
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize,
		TaskQueue:      make([]chan ziface.IRequset, utils.GlobalObject.WorkerPoolSize),
	}
}

func (mh *MsgHandle) DoMsgHandler(req ziface.IRequset) {
	handler, ok := mh.Apis[req.GetMsgID()] // 拿到对应的 handle
	if !ok {
		fmt.Printf("api msgID:[%d] is not found\n", req.GetMsgID())
		return
	}

	// 执行对应处理方法
	handler.PreHandle(req)
	handler.Handle(req)
	handler.PostHandle(req)
}

func (mh *MsgHandle) AddRouter(msgID uint32, router ziface.IRouter) {
	// 判断方法 msgID handle是否存在
	if _, ok := mh.Apis[msgID]; ok {
		panic("repeated api, msgID = " + strconv.Itoa(int(msgID)))
	}
	// 添加 handle
	mh.Apis[msgID] = router
	fmt.Println("Add api msgID = ", msgID)
}

// 一个 worker 工作流程
func (mh *MsgHandle) StartOneWorker(wokerID int, taskQueue chan ziface.IRequset) {
	fmt.Printf("Worker ID:[%d] is started\n", wokerID)
	for {
		select {
		case req := <-taskQueue:
			mh.DoMsgHandler(req)
		}
	}
}

// 工作池
func (mh *MsgHandle) StartWorkerPool() {
	// 遍历需要启动的 worker 的数量，依次启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个 worker 启动
		// 给任务队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequset, utils.GlobalObject.MaxWokerTaskLen)
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 添加任务
func (mh *MsgHandle) SendMsgToTaskQueue(req ziface.IRequset) {
	// 根据 ConnID 来分配当前连接请求由哪个 worker 负责处理 (0~9)
	workerID := req.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Printf("Add ConnID:[%d], req msgID:[%d] --> workerID:[%d]\n",
		req.GetConnection().GetConnID(), req.GetMsgID(), workerID)
	// 请求消息发送给任务队列
	mh.TaskQueue[workerID] <- req
}
