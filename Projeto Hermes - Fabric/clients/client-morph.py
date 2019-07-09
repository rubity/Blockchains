import sys
sys.path.insert(0, "..")
import math

import phe.encoding
from phe import paillier

import time
import array as arr
from opcua import Client as client_opcua
from hfc.fabric import Client as client_fabric

import json
import pickle

if __name__ == "__main__":

    #test if the meter ID was informed as argument
    if len(sys.argv) != 2:
        print("Usage:",sys.argv[0],"<meter id>")
        exit(1)

    #get the meter ID
    meter_id = sys.argv[1]

    #format the name of the expected public key
    pub_key_file = meter_id + ".pub"

    #try to retrieve the public key
    pub_key = pickle.load(open(pub_key_file, "rb"))

    #instantiate the hyperledeger fabric client
    c_hlf = client_fabric(net_profile="ptb-network.json")
    #defines the chaincode version
    cc_version = "1.0"

    #instantiate the opcua client
    c_opcua = client_opcua("opc.tcp://localhost:4840/freeopcua/server/")

    try:
        #get access to Fabric as Admin user
        admin = c_hlf.get_user('ptb.de', 'Admin')
        #the Fabric Python SDK do not read the channel configuration, we need to add it mannually
        c_hlf.new_channel('ptb-channel')

        #generates de pub_key in string format
        pub_key_str = "512," + str(pub_key.n) + "," + str(pub_key.g) + "," + str(pub_key.nsquare)

        #We invoke the chaincode 'registerMeter'. The transaction uses the meter ID '666' for
        #inserting the new measurement. User admin is used. 
        response = c_hlf.chaincode_invoke(requestor=admin,channel_name='ptb-channel',
                    peer_names=['peer0.ptb.de'],args=[meter_id,pub_key_str],
                    cc_name='fabmorph',cc_version=cc_version,fcn='registerMeter')

        #connect to the opcua server
        c_opcua.connect()
        #opcua client has a few methods to get proxy to UA nodes that should always be in address space such as Root or Objects
        root = c_opcua.get_root_node()
        #print shows what is happening
        print("OPC-UA Objects node is: ", root)

        #creates a accumulator to store the consumption locally
        local_consumption = 0
        read_control = 1

        while True:
            #use the this block of code to wait a little (we define our sampling rate)
            #or to ask for a key press from the user
            #time.sleep(5)
            input("Press ENTER to continue inserting...")

            #gets the measurement sample from opcua server
            sample = root.get_child(["0:Objects", "2:MyObject", "2:System Load"])
            
            #print shows what is happening
            print("Sampled measurement: ", sample.get_value())

            #inserts the individual measurement into var param
            measurement = int(sample.get_value() * 10)

            encrypted = str(pub_key.raw_encrypt(measurement))

            #invoke the LR chaincode... 
            print("Invoking insertMeasurement chaincode... ", encrypted)
                
            #the chaincode calls 'insertMeasurement'. The transaction uses the meter ID for
            #inserting the new measurement. User admin is used. 
            response = c_hlf.chaincode_invoke(requestor=admin,channel_name='ptb-channel',
                        peer_names=['peer0.ptb.de'],args=[meter_id,encrypted],
                        cc_name='fabmorph',cc_version=cc_version,fcn='insertMeasurement'
                        )
            #let's see what we did get...
            print(response)

            #increases the accumulator
            local_consumption += measurement
            # read_control += 1

            # if read_control > 4:
            #     #time to check values in blockchain
            #     response = c_hlf.chaincode_invoke(requestor=admin,channel_name='ptb-channel',
            #             peer_names=['peer0.ptb.de'],args=[meter_id],
            #             cc_name='fabmorph',cc_version=cc_version,fcn='getConsumption'
            #             )
            #     #response has the key/value struct in JSON format, so we use json library to load it
            #     data = json.loads(response)
            #     #get the encrypmeasure field and decrypt it
            #     decrypted = priv_key.raw_decrypt(int(data['encrypmeasure']))
                
            #     #show message comparing both values
            #     print("Comparing local and blockchain consumption: ",local_consumption,"/",decrypted)

            #     #reset read controller
            #     read_control = 1

            #so far, so good
            print("Insertion OK, getting next measurement (",local_consumption,")")

    finally:
        #only opcua client need to be disconnected
        c_opcua.disconnect()