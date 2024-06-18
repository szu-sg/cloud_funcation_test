import json
import marscode_baas_sdk.redis as redis

from dataclasses import dataclass
from typing import TypedDict, Optional, Any
from runtime import Args

@dataclass
class Input:
    """Function input input
    
    :var name: name of todo
    """
    name: str

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
    """Create a todo

    Notice: Each file needs to have a handler function as the entry function of the API.

    :param args: parameters of the entry function
    :param args.logger: logger instance used to print logs, injected by runtime
    :param args.input: parameters of http api, which can be parameters passed in query/body mode
    :returns: http api response
    """

    if args.input.name is None:
        return {
            "code": 1001,
            "message": "Input invalid, todo name is required",
            "data": None
        }

    args.logger.info("todo name is %s", args.input.name)

    result = redis.hset(
        "todo:list",
        args.input.name,
        json.dumps({
            "name": args.input.name,
            "state": 1
        })
    )

    # Return the body of api response
    if result == 1:
        return {
            "code": 0,
            "message": "OK",
            "data": None
        }
    else:
        return {
            "code": 1002,
            "message": f"Todo '{args.input.name}' already exists",
            "data": None
        }