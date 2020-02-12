import json


class ResultData(dict):
    def __init__(self, **kwargs):
        super(ResultData, self).__init__()
        self.delivered = 0
        self.released = 0
        self.rejected = 0
        self.modified = 0
        self.accepted = 0
        self.settled = 0
        self.errormsg = None
        for k in kwargs:
            getattr(self, k)
            setattr(self, k, kwargs[k])

    def __str__(self):
        return json.dumps(self.__dict__)
