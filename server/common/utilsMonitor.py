from . import utils
import threading

class MonitorUtils:
    def __init__(self):
        self._lock = threading.Lock()

    def store_bets_thread_safe(self, bets: list[utils.Bet]) -> None:
        with self._lock:
            utils.store_bets(bets)
    
    def load_winners_thread_safe(self, agent_id: int) -> list[utils.Bet]:
        with self._lock:
            winners = []
            for bet in utils.load_bets():
                if bet.agency == agent_id and utils.has_won(bet):
                    winners.append(int(bet.document))
            return winners