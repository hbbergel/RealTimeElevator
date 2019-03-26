package main


import "flag"
import "fmt"
import "os"
import "reflect"
import "time"

import "./bcast"
import "./elevio"
import "./elevstates"
import "./fsm"
import "./localip"
import "./peers"
import "./queue"
import "./types"




func main(){

    numFloors := 4

	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	var driver_port string
	flag.StringVar(&driver_port, "driver_port", "", "port to connecto to the elevator")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	
	}
	if driver_port == "" {
		driver_port = "20025"
	}

	fmt.Println(localip.LocalIP())

	elevio.Init("localhost:"+driver_port, numFloors)
	//fmt.Println("Stoplys av")

	

	// Channels

	peerUpdateCh            := make(chan peers.PeerUpdate)
	peerUpdateToAssigner    := make(chan peers.PeerUpdate)
	peerUpdateToLostPeers   := make(chan peers.PeerUpdate)
	peerTxEnable            := make(chan bool)

	orderDoneTx := make(chan types.Button)
	orderDoneRx := make(chan types.Button)

	// assignedOrder_netTx := make(chan types.Order)
	// assignedOrder_netRx := make(chan types.Order)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors  := make(chan int)
	//drv_motor_dir := make(chan elevio.MotorDirection)

	localState              := make(chan types.ElevState)
	allStates               := make(chan map[string]types.ElevState)
	allStatesToAssigner     := make(chan map[string]types.ElevState)
	allStatesToLostPeers    := make(chan map[string]types.ElevState)
	
	newOrder := make(chan types.Button, 1000)
	orderDone := make(chan types.Button)
	assignedOrder := make(chan types.Order)

	//portCh := make(chan string)

	//fsm_move := make(chan bool)



	//Goroutines


	go peers.Transmitter(20025, id, peerTxEnable)
	go peers.Receiver(20025, peerUpdateCh)

	go bcast.Transmitter(15000, orderDoneTx)
	go bcast.Receiver(15000, orderDoneRx)




	// go bcast.Transmitter(15001, assignedOrder_netTx)
	// go bcast.Receiver(15001, assignedOrder_netRx)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	go elevstates.ElevStates(id, localState, allStates)

	// TODO: Create lights goroutine here:
    //   Takes allStates, but this must be repeated (goes to lights and queue.Asssigner) 
    //   Sets all lights: hall for all elevs, cab for us 

    go Repeater(peerUpdateCh, nil, peerUpdateToLostPeers, peerUpdateToAssigner)
    go Repeater(allStates, nil, allStatesToLostPeers, allStatesToAssigner)
    
	go queue.Assigner(id, drv_buttons, allStatesToAssigner, peerUpdateToAssigner, assignedOrder)
	go queue.LostPeers(peerUpdateToLostPeers, allStatesToLostPeers, newOrder)
	go queue.Distributor(id, assignedOrder, newOrder)

	go fsm.Fsm_run_elev(newOrder, drv_floors, orderDone, localState)


	//go ReInitializing()
	
    
    
    for {
        select {
            
		case a := <- orderDone:
			fmt.Printf("Order done: %+v\n", a)
			if a.Type == 0 || a.Type == 2 || a.Floor == 0 {
				elevio.SetButtonLamp(0, a.Floor, false)
				elevio.SetButtonLamp(2, a.Floor, false)
			} else if a.Type == 1 || a.Type == 2 || a.Floor == 3 {
				elevio.SetButtonLamp(2, a.Floor, false)
				elevio.SetButtonLamp(1, a.Floor, false)
			}
			ticker := time.NewTimer(time.Millisecond)

			go func() {
				for{
					select{
					case <- ticker.C:
						orderDoneTx <- a
						ticker.Stop()
					}
				}
			}()	
		case a := <- orderDoneRx:
			ClearHallLights(a)

	




		}
	}
	    
}


/*
Repeats values sent on an input channel to several output channels
    ch_in:      The input channel, of type 'T'
    dup_fn:     A function of type 'fn(T) T' that creates a deep copy, 
                or 'nil' if a shallow copy is sufficient
    chs_out:    Several output channels, all of type 'T'
*/
func Repeater(ch_in interface{}, dup_fn interface{}, chs_out ...interface{}) {
	T := reflect.TypeOf(ch_in).Elem()
	for n, c := range chs_out {
		T_out := reflect.TypeOf(c).Elem()
		if T_out != T {
			panic(fmt.Sprintf("All channels must be of the same type. Got '%v' as input channel, and '%v' as output channel number %v", T, T_out, n+1))
		}
	}
	if dup_fn != nil {
		F := reflect.TypeOf(dup_fn)
		if !((F.Kind() == reflect.Func) && (F.NumIn() == 1) && (F.NumOut() == 1) && (F.In(0) == T) && (F.Out(0) == T)) {
			panic(fmt.Sprintf("Duplication function must be 'nil' or of the type 'func(%v) %v' (got '%v')", T, T, F))
		}
	}
	for {
		v, _ := reflect.ValueOf(ch_in).Recv()

		v2 := reflect.New(T)
		if dup_fn != nil {
			v2 = reflect.ValueOf(dup_fn).Call([]reflect.Value{v})[0]
		} else {
			v2 = v
		}

		for _, c := range chs_out {
			reflect.ValueOf(c).Send(v2)
		}
	}
}


func ClearHallLights(a types.Button)  {
	elevio.SetButtonLamp(elevio.ButtonType(a.Type), a.Floor, false)
	// if a.Type == 0 || a.Floor == 0 {
    //     elevio.SetButtonLamp(0, a.Floor, false)
	// } else if a.Type == 1 || a.Floor == 3 {
    //     elevio.SetButtonLamp(1, a.Floor, false)
    // } 
}
