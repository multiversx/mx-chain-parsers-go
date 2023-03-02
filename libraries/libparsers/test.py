import ctypes
import json
import platform
import pprint
import urllib.request
from pathlib import Path
from typing import Any, Dict, Optional


class LibParsersFacade:
    _library: Optional[ctypes.CDLL] = None

    def __init__(self) -> None:
        self.indexed_transaction_parser_handle = self._new_indexed_transaction_parser({
        })

    def _new_indexed_transaction_parser(self, config: Any) -> int:
        func = self._get_library().newIndexedTransactionParser

        input_json = json.dumps(config).encode()
        output = func(input_json)
        output_typed = ctypes.c_int(output)
        return int(output_typed.value)

    def parse_indexed_transaction(self, transaction: Any) -> Dict[str, Any]:
        func = self._get_library().parseIndexedTransaction

        input_json = json.dumps(transaction).encode()
        output = func(self.indexed_transaction_parser_handle, input_json)
        output_typed = ctypes.string_at(output)
        output_json = output_typed.decode()
        output_dict = json.loads(output_json)

        return output_dict

    def _get_library(self) -> ctypes.CDLL:
        if self._library is None:
            self._library = self._load_library()

        return self._library

    def _load_library(self) -> ctypes.CDLL:
        lib_path = self._get_library_path()

        if not lib_path.exists():
            raise Exception(f"cannot load library: {lib_path}")

        lib = ctypes.CDLL(str(lib_path), winmode=0)

        lib.newIndexedTransactionParser.argtypes = [ctypes.c_char_p]
        lib.newIndexedTransactionParser.restype = ctypes.c_int

        lib.parseIndexedTransaction.argtypes = [ctypes.c_int, ctypes.c_char_p]
        lib.parseIndexedTransaction.restype = ctypes.c_char_p

        print(f"Loaded library: {lib_path}")

        return lib

    def _get_library_path(self) -> Path:
        os_name = platform.system()

        if os_name == "Darwin":
            lib_name = "libparsers.dylib"
        elif os_name == "Linux":
            lib_name = "libparsers.so"
        else:
            raise Exception(f"unsupported operating system: {os_name}")

        return (Path(__file__).parent / lib_name).resolve()


def main():
    facade = LibParsersFacade()
    address = "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th"
    url = f"https://devnet-api.multiversx.com/accounts/{address}/transfers?size=5"
    request = urllib.request.Request(url)
    response = urllib.request.urlopen(request)
    data = json.loads(response.read())

    for transaction in data:
        parsed = facade.parse_indexed_transaction(transaction)
        pprint.pprint(parsed)


if __name__ == "__main__":
    main()
