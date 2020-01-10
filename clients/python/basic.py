import math
import signal
import sys
import time
import logging
from proton.handlers import MessagingHandler
from proton.reactor import Container

from config import Config
from result import ResultData


class BasicCommon(MessagingHandler):
    def __init__(self):
        super(BasicCommon, self).__init__(auto_accept=True, auto_settle=True)

        # Common
        self.error: bool = False
        self.result_data: ResultData = ResultData()

        # Container
        self._container = None

        # Configuration
        self._config = Config("basic")

        # DEFAULT config
        self._url = self._config.get("AMQP_URL")
        self._timeout = int(self._config.get("AMQP_TIMEOUT"))

        # Specialized config
        self._msgcount = int(self._config.get("MSG_COUNT"))
        self._msgsize = int(self._config.get("MSG_SIZE"))
        self._msgpattern = self._config.get("MSG_PATTERN")

        self._expected_body = BasicCommon.generate_message_body(self._msgsize, self._msgpattern)

    def timedout(self, signum, frame):
        logging.debug("timed out")
        self.set_error("Timed out")
        raise TimeoutError()

    def set_error(self, msg: str):
        self.error = True
        self.result_data.errormsg = msg

    def execute_client(self):
        """
        This method must be executed as the main method in the concrete client.
        :return:
        """
        error = False
        try:
            signal.signal(signal.SIGALRM, self.timedout)
            signal.alarm(self._timeout)
            self._container = Container(self)
            self._container.run()
        except TimeoutError as e:
            if self._container:
                self._container.stop()
            error = True
        except Exception as e:
            logging.debug("Exception caught: %s" % e)
            self.set_error(e.__str__())
            error = True
        finally:
            signal.signal(signal.SIGALRM, signal.SIG_IGN)
            time.sleep(5)
            # This must be the last thing in the output (which will be parsed by test suite)
            print(self.result_data, flush=True)
            sys.exit(1 if error else 0)

    @staticmethod
    def generate_message_body(size, pattern) -> str:
        return (pattern * math.ceil(size / len(pattern)))[:size]
