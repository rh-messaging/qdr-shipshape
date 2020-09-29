from proton import Message
from proton.reactor import AtLeastOnce
from .basic import BasicCommon
import uuid
import sys
import logging


class BasicSender(BasicCommon):
    def __init__(self):
        super(BasicSender, self).__init__()
        # Sender specifics
        self._sender = None
        self._released_tags = []
        self._multicast = 'multicast' in self._url

    #
    # Internal methods
    #
    def can_send(self) -> bool:
        rd = self.result_data
        # Must be done first (to block a possible retry in case done sending)
        if self.done_sending():
            return False

        # when multicast always send
        if self._multicast:
            return True

        # Number of pending acks (sent - released - accepted)
        pendingacks = rd.delivered - rd.released - rd.accepted

        # Proceed only if accepted + pending < total
        # This avoids continue sending when we have pending acks
        return rd.accepted + pendingacks < self._msgcount

    def done_sending(self) -> bool:
        return self.result_data.accepted == self._msgcount

    #
    # Event handling
    #
    def on_start(self, event):
        self._sender = event.container.create_sender(self._url)

    def on_sendable(self, event):
        self.send(event, 'on_sendable')

    def send(self, event, source):
        if not self._sender.credit or not self.can_send():
            print("[%s] unable to send - credit: %s - partial results: %s" % (source, self._sender.credit, self.result_data))
            logging.debug("[%s] unable to send - credit: %s - partial results: %s" % (source, self._sender.credit, self.result_data))
            # retry in 1 sec
            if not self.done_sending():
                event.reactor.schedule(1, self)
            return
        logging.debug("[%s] message sent: credit: %s - partial results: %s" % (source, self._sender.credit, self.result_data))
        msg = Message(id=str(uuid.uuid1()), body=self._expected_body)
        self._sender.send(msg)
        self.result_data.delivered += 1

    def on_timer_task(self, event):
        self.send(event, "on_timer_task")

    def on_accepted(self, event):
        logging.debug("message accepted: %s" % event.delivery.tag)
        self.result_data.accepted += 1

        if not self._multicast and self.done_sending():
            self.done(event)

    def done(self, event):
            logging.debug("done sending")
            event.sender.close()
            event.connection.close()

    def on_released(self, event):
        # this is to prevent an issue we faced with two on_released
        # calls happening for same delivery tag
        # related proton frames below:
        #
        # [0x562a0083ed80]:0 <- @disposition(21) [role=true, first=981, state=@released(38) []]
        # [0x562a0083ed80]:0 <- @disposition(21) [role=true, first=981, last=982, settled=true, state=@released(38) []]
        #
        # in the sample above, the on_released was invoked 3 times for: 981, 981 and 982.
        if event.delivery.tag in self._released_tags:
            return
        self._released_tags.append(event.delivery.tag)
        logging.debug("message released: %s" % event.delivery.tag)
        self.result_data.released += 1
        self.send(event, 'on_released')

    def on_rejected(self, event):
        logging.debug("message rejected: %s" % event.delivery.tag)
        self.result_data.rejected += 1
        self.send(event, 'on_released')

    def on_settled(self, event):
        logging.debug("message settled: %s" % event.delivery.tag)
        self.result_data.settled += 1


if __name__ == "__main__":
    BasicSender().execute_client()
