from dataclasses import dataclass
from typing import TypedDict, Optional, Any
from runtime import Args

@dataclass
class Input:
    """Function input params
    
    :var name: name of user
    """
    name: Optional[str]

class Output(TypedDict):
    """Function output data
    
    :var message: reply to greet the user
    """
    message: str

def handler(args: Args[Input]) -> Output:
    """Say hello to the user when he introduces himself"""

    name = args.input.name if hasattr(args.input, 'name') and args.input.name is not None else 'world'

    args.logger.info("user name is %s", name)

    return {
        "message": f"Hello {name}"
    }

