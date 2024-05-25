import requests

class KVRestAPI:
	def __init__(self, api_key, base_url="https://kvrest.dev/api"):
		self.api_key = api_key
		self.base_url = base_url
		self.headers = {"API-KEY": self.api_key, "Content-Type": "application/json"}

	def _make_request(self, method, endpoint, json=None):
		url = f"{self.base_url}{endpoint}"
		response = requests.request(method, url, headers=self.headers, json=json)

		if response.status_code not in (200, 201):
			response.raise_for_status()  # Raise an exception for bad status codes

		try:
			if response.text:  # Check if response has content
				return response.json()
			else:
				return {}  # Or return None, depending on how you want to handle it
		except ValueError as e: 
			# Log the error and the raw response for debugging
			print(f"Error parsing JSON response: {e}, Raw Response: {response.text}")
			return None # Or raise the exception again if needed

	def create_bucket(self, bucket_name):
		"""Create a new bucket."""
		endpoint = f"/{bucket_name}"
		self._make_request("PUT", endpoint)

	def delete_bucket(self, bucket_name):
		"""Delete an existing bucket."""
		endpoint = f"/{bucket_name}"
		self._make_request("DELETE", endpoint)

	def list_buckets(self):
		"""List all buckets."""
		endpoint = "/buckets"
		return self._make_request("POST", endpoint)

	def create_or_update_key_value(self, bucket_name, key, value):
		"""Create or update a key-value pair."""
		endpoint = f"/{bucket_name}/{key}"
		self._make_request("PUT", endpoint, json=value)

	def get_value(self, bucket_name, key):
		"""Retrieve a value for a key."""
		endpoint = f"/{bucket_name}/{key}"
		return self._make_request("GET", endpoint)

	def delete_key_value(self, bucket_name, key):
		"""Delete a key-value pair."""
		endpoint = f"/{bucket_name}/{key}"
		self._make_request("DELETE", endpoint)

	def list_keys(self, bucket_name):
		"""List all keys in a bucket."""
		endpoint = f"/{bucket_name}"
		return self._make_request("GET", endpoint)


if __name__ == "__main__":
	api_key = "CHANGE_ME"  # Replace with your actual API key
	api = KVRestAPI(api_key)

	# Example usage:
	try:
		bucket_name = "my-test-bucket"
		key = "my-key"
		value = {"message": "Hello from the API!"}

		# Create a bucket
		api.create_bucket(bucket_name)
		print(f"Bucket '{bucket_name}' created successfully.")

		# Create a key-value pair
		api.create_or_update_key_value(bucket_name, key, value)
		print(f"Key-value pair created/updated for key '{key}'.")

		# Retrieve the value
		retrieved_value = api.get_value(bucket_name, key)
		print(f"Retrieved value for key '{key}': {retrieved_value}")

		# List all buckets
		buckets = api.list_buckets()
		print(f"All buckets: {buckets}")

		# List keys in the bucket
		keys = api.list_keys(bucket_name)
		print(f"Keys in bucket '{bucket_name}': {keys}")

		# Delete the key-value pair
		api.delete_key_value(bucket_name, key)
		print(f"Key-value pair with key '{key}' deleted.")

		# Delete the bucket
		api.delete_bucket(bucket_name)
		print(f"Bucket '{bucket_name}' deleted.")

	except requests.exceptions.RequestException as e:
		print(f"An error occurred: {e}")