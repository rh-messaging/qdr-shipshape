from basic import BasicCommon
import sys
import time
import logging


class BasicReceiver(BasicCommon):
    def __init__(self):
        super(BasicReceiver, self).__init__()
        # Receiver specifics
        self._receiver = None
        self._last_received_id = None
        self._disconnected = False

    #
    # Event handling
    #
    def on_start(self, event):
        self._receiver = event.container.create_receiver(self._url)

    def on_message(self, event):
        # ignore if received a dup
        if event.message.id and event.message.id == self._last_received_id:
            logging.debug("releasing duplicated message")
            self.release(event.delivery)
            return

        if event.message.id:
            self._last_received_id = event.message.id

        if self.result_data.delivered < self._msgcount:
            self.result_data.delivered += 1
            self.accept(event.delivery)
            event.delivery.settle()
            logging.debug("message received: %s" % self._last_received_id)

            # If body does not match expected value
            if event.message.body != self._expected_body:
                self.set_error("Invalid message body received. Expected: %s. Got: %s"
                               % (self._expected_body, event.message.body))
                self._disconnect(event)

            if self.result_data.delivered == self._msgcount:
                self._disconnect(event)
        else:
            logging.debug("received enough, releasing")
            self.release(event.delivery)
            self._disconnect(event)

    def _disconnect(self, event):
        if self._disconnected:
            return
        logging.debug("disconnecting")
        self._disconnected = True
        if event.receiver:
            event.receiver.detach()
            event.receiver.close()
        if event.connection:
            event.connection.close()
        


if __name__ == "__main__":
    BasicReceiver().execute_client()
