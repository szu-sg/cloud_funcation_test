import json
import marscode_baas_sdk.redis as redis


from dataclasses import dataclass
from typing import TypedDict, Optional, Any
from runtime import Args

@dataclass
class Input:
    """Function input input

    :var state: state of todo, 1: active, 2: completed,
    """
    state: Optional[int]

class Output(TypedDict):
    """Function output data
    
    :var code: status code of output
    :var message: this is the message of output
    :var data: this is the data of output
    """
    code: int
    message: str
    data: Optional[dict[str, Any]]

def hgetall(key):
    return dict(redis.hgetall(key))

def filter_todos(todos, state=None):
    if state is None:
        return todos
    return [todo for todo in todos if todo['state'] == state]

def handler(args: Args[Input]) -> Output:
    """Search todos

    Notice: Each file needs to have a handler function as the entry function of the API.

    :param args: parameters of the entry function
    :param args.logger: logger instance used to print logs, injected by runtime
    :param args.input: parameters of http api, which can be parameters passed in query/body mode
    :returns: http api response
    """

    result = hgetall('todo:list')

    data = [json.loads(todo) for todo in result.values()]

    # Filter todo items based on status field
    if hasattr(args.input,'state'):
        return {'code': 0, 'message': 'ok', 'data': filter_todos(data, args.input.state)}
    return {'code': 0, 'message': 'ok', 'data': data}
    