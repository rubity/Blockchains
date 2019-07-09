from hfc.fabric import Client as client_fabric
from tornado.platform.asyncio import AnyThreadEventLoopPolicy
import json
from waitress import serve
import asyncio
from flask import *

'''Defines the chaincode version'''
cc_version = "1.0"

'''Necessary to run Flask'''
app = Flask(__name__)

'''To solve a problem with asyncio'''
asyncio.set_event_loop_policy(AnyThreadEventLoopPolicy())

@app.route('/', methods=['POST'])
def insertBlockchain():

    loop = asyncio.get_event_loop()

    meter_id = "666"
    '''CHANGE THIS'''

    '''instantiate the hyperledger fabric client'''
    c_hlf = client_fabric(net_profile="ptb-network2.json")

    '''get access to Fabric as Admin user'''
    admin = c_hlf.get_user('ptb.de', 'Admin')

    '''the Fabric Python SDK do not read the channel configuration, we need to add it mannually'''
    c_hlf.new_channel('ptb-channel')

    '''Receive the data'''
    data = request.data

    '''Convert in a Json structure'''
    parsed = json.loads(data)
    print(json.dumps(parsed, indent=4, sort_keys=True))

    '''Insert in the Blockchain'''
    response = loop.run_until_complete(c_hlf.chaincode_invoke(requestor=admin, channel_name='ptb-channel', peers=['peer0.ptb.de'],
                                             args=[meter_id, data], cc_name='fabmorph', cc_version=cc_version,
                                             fcn='insertMeasurement', cc_pattern=None))

    print(response)
    return 'Okay, the data was stored'

'''Defines the the host and port of communication'''
serve(app, host='0.0.0.0', port=8080)

if __name__ == "__main__":
    app.run(debug=True)






