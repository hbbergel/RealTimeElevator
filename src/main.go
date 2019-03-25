package main

import "./elevio"
import "./queue"
import "./types"
import "./elevstates"
import "fmt"
import "./bcast"
import "./localip"
import "./peers"
import "flag"
import "os"



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
		driver_port = "15657"
	}

	var all_states map[string]types.ElevState

	// Channels

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	peerList := make(chan []string)

	buttonTx := make(chan types.Button)
	buttonRx := make(chan types.Button)

	assignedOrder_netTx := make(chan types.Order)
	assignedOrder_netRx := make(chan types.Order)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors  := make(chan int)
	drv_motor_dir := make(chan elevio.MotorDirection)

	local_state := make(chan types.ElevState)
	all_statesCh := make(chan map[string]types.ElevState)
	
	newOrder := make(chan types.Order)
	orderDone := make(chan bool)
	assignedOrder := make(chan types.Order)

	portCh := make(chan string)

	fsm_move := make(chan bool)



	//Goroutines

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	go func(peerUpdateCh <-chan peers.PeerUpdate, peerList chan<- []string) {
		p := <- peerUpdateCh
		for {
			select {
			case peerList <- p.Peers:
			}
		}
	}()	

	go bcast.Transmitter(16569, buttonTx)
	go bcast.Receiver(16569, 	buttonRx)

	go bcast.Transmitter(15002, assignedOrder_netTx)
	go bcast.Receiver(15002, assignedOrder_netRx)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)

	go elevstates.ElevStates(id, local_state, all_statesCh)

	//go queue.UpdateQueue(drv_buttons, all_statesCh, peerUpdateCh, updateOrder)
	go queue.Assigner(id, drv_buttons, all_statesCh, peerList, assignedOrder)
			
	elevio.Init("localhost:"+driver_port, numFloors)


	go func(assignedOrder_netRx <-chan types.Order, newOrder chan<- types.Order){
		for {
			select{
			case order := <- assignedOrder_netRx:
			if order.AssignedTo == id {
				newOrder <- order
				}
			}	
		}
	}()
    
    
    for {
        select {
        case a := <- drv_buttons:
			fmt.Printf("drv_buttons: %#v\n", a)
			buttonTx <- types.Button{Floor: a.Floor, Type: int(a.Button)}
					
		            
		case a := <- drv_floors:			
			fmt.Printf("drv_floors:  %#v\n", a)
		
			
            
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <-buttonRx:
			fmt.Printf("buttonRx:    %#v\n", a)
		}
    }    
}
