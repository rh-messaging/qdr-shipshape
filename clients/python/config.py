#
# Configuration parser
#

import configparser
import os
import logging
from abc import ABC


class Config(ABC):
    """
    Config provides mechanism to load the config.ini file from
    CLIENT_CONFIG_DIR environment variable, or local directory
    to the python script.

    The options defined in the property can be overwritten via
    environment variables, so it is important to avoid clashing
    on the option names.

    Options are also organized in sections, and each section is
    named as the python file being executed (string at the left
    of first dot). Or it can be defined by "clientid"
    initialization argument.
    """
    def __init__(self, clientid: str = ""):
        """
        Loads the config.ini and includes both the DEFAULT plus the
        specialized section, which is determined by the python file
        name in use (default), left string before first dot. Or the
        section can be specified through the clientid init argument.
        :param clientid:
        """
        self.config = configparser.ConfigParser()
        self._clientid = clientid
        # can determine client section to use based on provided clientid or python script filename
        self._clientid_section = (clientid or os.path.basename(__file__)).split(os.extsep)[0]
        self._config_dir = os.getenv("CLIENT_CONFIG_DIR", os.path.dirname(os.path.abspath(__file__)))
        self._config_file = os.path.join(self._config_dir, "config.ini")
        # If config cannot be determined, fail
        if not os.path.exists(self._config_file):
            raise Exception("Config file could not be found: %s" % self._config_file)
        self.config.read(self._config_file)

        # initialize logging
        loglevel = "DEBUG" if os.getenv("PN_TRACE_FRM", "0") == "1" else "ERROR"
        logging.basicConfig(level=loglevel,
            format='%(asctime)s [%(levelname)s] (%(filename)s:%(lineno)s) - %(message)s')


    def get_default(self, option: str, _default: str = "") -> str:
        """
        Returns the given option from the ENVIRONMENT variable dict or
        from the DEFAULT section of the ini file. If none found,
        then the _default is returned (if provided) or None.
        :param option:
        :param _default:
        :return:
        """
        return os.getenv(option, self.config.get("DEFAULT", option, fallback=_default))

    def get_specialized(self, option: str, _default: str = "") -> str:
        """
        Returns the given option from the ENVIRONMENT variable dict or
        from the specialized section (based on script name or clientid).
        If nothing found, it returns the _default or None.
        :param option:
        :param _default:
        :return:
        """
        return os.getenv(option, self.config.get(self._clientid_section, option, fallback=_default))

    def get(self, option: str, _default: str = ""):
        """
        Retrieves a given option from the ENVIRONMENT dict, from
        the DEFAULT section or from the specialized section.
        If none available, then returns the _default (if provided) or None.
        :param option:
        :param _default:
        :return:
        """
        return self.get_default(option, _default) or self.get_specialized(option, _default)


if __name__ == "__main__":
    c = Config('basic')
    print(c.get('AMQP_URL'))
    print(c.get('MSG_SIZE'))
    print(c.get('MSG_COUNT'))
    print(c.get('MSG_PATTERN'))
