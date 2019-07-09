from hfc.fabric import Client as client_fabric
import asyncio

#defines the chaincode version
cc_version = "1.0"

def main():

    teste()

def teste():

    loop = asyncio.get_event_loop()

    meter_id = "666"
    measurement = "123"

    '''instantiate the hyperledger fabric client'''
    c_hlf = client_fabric(net_profile="ptb-network.json")

    '''get access to Fabric as Admin user'''
    admin = c_hlf.get_user('ptb.de', 'Admin')

    # Query Peer installed chaincodes, make sure the chaincode is installed
    response = loop.run_until_complete(c_hlf.query_installed_chaincodes(
        requestor=admin,
        peers=['peer0.ptb.de'],
        decode=True
    ))

    print(response)


    '''the Fabric Python SDK do not read the channel configuration, we need to add it mannually'''
    c_hlf.new_channel('ptb-channel')

    response = loop.run_until_complete(c_hlf.chaincode_invoke(requestor=admin, channel_name='ptb-channel', peers=['peer0.ptb.de'],
                                            args=[meter_id], cc_name='fabmorph', cc_version=cc_version,
                                            fcn='queryHistory', cc_pattern=None))

    print(response)

if __name__ == "__main__":
    main()

