import json
import marscode_baas_sdk.redis as redis

from dataclasses import dataclass
from typing import TypedDict, Optional, Any
from runtime import Args

@dataclass
class Input:
    """Function input input
    
    :var name: name of todo
    :var state: state of todo, 1: active, 2: completed,
    """
    name: str
    state: int

class Output(TypedDict):
    """Function output data
    
    :var code: status code of output
    :var message: this is the message of output
    :var data: this is the data of output
    """
    code: int
    message: str
    data: Optional[dict[str, Any]]

def handler(args: Args[Input]) -> Output:
    """Edit a todo

    Notice: Each file needs to have a handler function as the entry function of the API.

    :param args: parameters of the entry function
    :param args.logger: logger instance used to print logs, injected by runtime
    :param args.input: parameters of http api, which can be parameters passed in query/body mode
    :returns: http api response
    """

    # Extract parameters
    name, state = args.input.name, args.input.state

    # Check if the parameters are valid
    if not name:
        return {'code': 1001, 'message': 'input invalid, todo name is required', 'data': None}
    if not state:
        return {'code': 1001, 'message': 'input invalid, todo state is required', 'data': None}
    if state not in [1,2]:
        return {'code': 1001,'message': 'input invalid, todo state value is invalid','data': state}

    args.logger.info('todo name is %s ,state is %s', name, state)
    redis.hset(
        'todo:list',
        name, 
        json.dumps({
            "name": name,
            "state": state
        }))

    # Return the body of api response
    return { 'code': 0, 'message': 'ok', 'data': None }