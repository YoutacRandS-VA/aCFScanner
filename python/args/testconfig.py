import argparse
import json
import os

from utils.exceptions import TemplateReadError, BinaryNotFoundError
from utils.os import detect_system
from xray import templates
from xray.binary import download_binary
from report.clog import CLogger
from utils.requests import download_file

logger = CLogger("testconfig")

PATH = os.path.dirname(os.path.abspath(__file__))
PARENT_PATH = os.path.dirname(PATH)


class TestConfig:
    @classmethod
    def from_args(cls, args: argparse.Namespace):
        """creates a TestConfig object from the arguments passed to the program

        Args:
            args (argparse.Namespace): arguments passed to the program
        """
        # create test config
        test_config = cls()
        
        # load config if need be
        if not args.no_vpn and args.template_path is None:
            if args.config_path is None:
                os.makedirs(os.path.join(PATH, ".tmp"), exist_ok=True)
                download_file(
                    url="https://raw.githubusercontent.com/MortezaBashsiz/CFScanner/main/bash/ClientConfig.json",
                    savepath=os.path.join(PATH, ".tmp", "sudoer_config.json")
                )
                args.config_path = os.path.join(PATH, ".tmp", "sudoer_config.json")
            with open(args.config_path, "r") as infile:
                jsonfilecontent = json.load(infile)
                test_config.user_id = jsonfilecontent["id"]
                test_config.ws_header_host = jsonfilecontent["host"]
                test_config.address_port = int(jsonfilecontent["port"])
                test_config.sni = jsonfilecontent["serverName"]
                test_config.user_id = jsonfilecontent["id"]
                test_config.ws_header_path = "/" + \
                    (jsonfilecontent["path"].lstrip("/"))



        if args.template_path is None:
            test_config.proxy_config_template = templates.vmess_ws_tls
        else:
            try:
                with open(args.template_path, "r") as infile:
                    test_config.proxy_config_template = infile.read()
            except FileNotFoundError:
                raise TemplateReadError("template file not found")
            except IsADirectoryError:
                raise TemplateReadError(
                    "template file is a directory. please provide the path to the file")
            except PermissionError:
                raise TemplateReadError(
                    "permission denied while reading template file")
            except Exception as e:
                raise TemplateReadError(
                    "error while reading template file: " + str(e))

        # speed related config
        test_config.startprocess_timeout = args.startprocess_timeout
        test_config.do_upload_test = args.do_upload_test or args.min_ul_speed is not None
        test_config.min_ul_speed = args.min_ul_speed or 50
        test_config.min_dl_speed = args.min_dl_speed
        test_config.max_dl_time = args.max_dl_time
        test_config.max_ul_time = args.max_ul_time
        test_config.fronting_timeout = args.fronting_timeout
        test_config.max_dl_latency = args.max_dl_latency
        test_config.max_ul_latency = args.max_ul_latency
        test_config.n_tries = args.n_tries
        test_config.novpn = args.no_vpn

        system_info = detect_system()
        if test_config.novpn:
            test_config.binpath = None
        elif args.binpath is not None:
            # Check if file exists
            if not os.path.isfile(args.binpath):
                raise BinaryNotFoundError(
                    "The binary path provided does not exist"
                )
            test_config.binpath = args.binpath
        else:
            test_config.binpath = download_binary(
                system_info=system_info,
                bin_dir=PARENT_PATH
            )


        return test_config
