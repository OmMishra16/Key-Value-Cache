import string
from locust import HttpUser, task, FastHttpUser, constant, events
import random
import uuid

# Configuration
KEY_POOL_SIZE = 5_000  # Reduced pool size
VALUE_LENGTH = 128    # Smaller values
PUT_RATIO = 0.3      # More reads than writes

class CacheUser(FastHttpUser):
    # Add small wait time to prevent overwhelming
    wait_time = constant(0.01)  

    # Pre-generate pools with fixed lengths
    key_pool = [f"key_{i:05d}" for i in range(KEY_POOL_SIZE)]
    value_pool = [''.join(random.choices(string.ascii_letters, k=VALUE_LENGTH))
                  for _ in range(100)]

    def on_start(self):
        """Initialize some data"""
        # Warm up the cache
        for i in range(min(100, KEY_POOL_SIZE)):
            self.client.post("/put",
                json={"key": self.key_pool[i], 
                      "value": random.choice(self.value_pool)},
                catch_response=True)

    @task(7)
    def get_request(self):
        key = random.choice(self.key_pool)
        with self.client.get(
            f"/get?key={key}",
            catch_response=True
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Status code: {response.status_code}")

    @task(3)
    def put_request(self):
        key = random.choice(self.key_pool)
        value = random.choice(self.value_pool)
        with self.client.post(
            "/put",
            json={"key": key, "value": value},
            catch_response=True
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Status code: {response.status_code}")