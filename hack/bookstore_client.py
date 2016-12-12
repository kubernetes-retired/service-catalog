#!/usr/bin/env python

# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


"""A simple client for the bookstore example application.

Steps:

1) Deploy bookstore application.
   HOST=deployed_host_name
   To use https protocol, HOST should start with https://
   Otherwise http protocol will be used.
2) Run:
   ./bookstore_client.py --host=$HOST --api_key=$KEY
"""

import argparse
import httplib
import json
import ssl
import sys


HTTPS_PREFIX = 'https://'
HTTP_PREFIX = 'http://'


class FLAGS:
    pass


def http_connection(host):
  if host.startswith(HTTPS_PREFIX):
      host = host[len(HTTPS_PREFIX):]
      print 'Use https to connect: %s' % host
      return httplib.HTTPSConnection(host)
  else:
      if host.startswith(HTTP_PREFIX):
          host = host[len(HTTP_PREFIX):]
      else:
          host = host
      print 'Use http to connect: %s' % host
      return httplib.HTTPConnection(host)


class Response(object):
    """A class to wrap around httplib.response class."""

    def __init__(self, r):
        self.text = r.read()
        self.status_code = r.status
        self.headers = r.getheaders()
        self.content_type = r.getheader('content-type')
        if self.content_type != None:
            self.content_type = self.content_type.lower()

    def json(self):
        try:
            return json.loads(self.text)
        except ValueError as e:
            print 'Error: failed in JSON decode: %s' % self.text
            return {}

    def is_json(self):
        if self.content_type != 'application/json':
            return False
        try:
            json.loads(self.text)
            return True
        except ValueError as e:
            return False


class BookstoreClient(object):
    def __init__(self, verify=False):
        self.conn = http_connection(FLAGS.host)
        self._verify = verify

    def assertEqual(self, a, b):
        if not self._verify:
            return

        msg = 'assertEqual(%s, %s)' % (str(a), str(b))
        if a != b:
            sys.exit('Equality assertion failed: %s, %s' % (str(a), str(b)))

    def _call_http(self, path, api_key=None, data=None, method=None):
        """Makes a http call and returns its response."""
        url = path
        headers = {'Content-Type': 'application/json'}
        if api_key:
            headers['x-api-key'] = api_key
        body = json.dumps(data) if data else None
        if not method:
            method = 'POST' if data else 'GET'
        if FLAGS.verbose:
            print 'HTTP: %s %s' % (method, url)
            print 'headers: %s' % str(headers)
            print 'body: %s' % body
        self.conn.request(method, url, body, headers)
        response = Response(self.conn.getresponse())
        if FLAGS.verbose:
            print 'Status: %s, body=%s' % (response.status_code, response.text)
        return response

    def _send_request(self, path, api_key=None,
                      data=None, method=None):
        if api_key:
            print 'Negative test: remove api_key.'
            r = self._call_http(path, None, data, method)
            self.assertEqual(r.status_code, 401)
            self.assertEqual(
                r.json()['message'],
                ('Method doesn\'t allow unregistered callers (callers without '
                 'established identity). Please use API Key or other form of '
                 'API consumer identity to call this API.'))
            print 'Completed unregistered test.'
            print 'Negative test: pass blocked api_key.'
            r = self._call_http(path, 'aaaa', data, method)
            self.assertEqual(r.status_code, 403)
            self.assertEqual(
                r.json()['message'], 'Client application blocked.')
            print 'Completed blocked api_key test.'
        return self._call_http(path, api_key, data, method)

    def clear(self):
        print 'Clear existing shelves.'
        response = self._send_request('/shelves')
        self.assertEqual(response.status_code, 200)
        json_ret = response.json()
        for shelf in json_ret.get('shelves', []):
            self.delete_shelf(shelf)

    def create_shelf(self, shelf):
        print 'Create shelf: %s' % str(shelf)
        # create shelves: api_key.
        response = self._send_request(
            '/shelves', api_key=FLAGS.api_key, data=shelf)
        self.assertEqual(response.status_code, 200)
        # shelf name generated in server, not the same as required.
        json_ret = response.json()
        self.assertEqual(json_ret.get('theme', ''), shelf['theme'])
        return json_ret

    def verify_shelf(self, shelf):
        print 'Verify shelf: shelves/%d' % shelf['id']
        # Get shelf: api_key.
        r = self._send_request('/shelves/%d' % shelf['id'], api_key=FLAGS.api_key)
        self.assertEqual(r.status_code, 200)
        self.assertEqual(r.json(), shelf)

    def delete_shelf(self, shelf):
        shelf_name = 'shelves/%d' % shelf['id']
        print 'Remove shelf: %s' % shelf_name
        # delete shelf: api_key
        r = self._send_request(
            '/' + shelf_name, api_key=FLAGS.api_key, method='DELETE')
        self.assertEqual(r.status_code, 204)

    def verify_list_shelves(self, shelves):
        # list shelves: no api_key
        response = self._send_request('/shelves')
        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json().get('shelves', []), shelves)

    def create_book(self, shelf, book):
        print 'Create book in shelf: %s, book: %s' % (shelf['id'], str(book))
        # Create book: api_key
        response = self._send_request(
            '/shelves/%d/books' % shelf['id'], api_key=FLAGS.api_key, data=book)
        self.assertEqual(response.status_code, 200)
        # book name is generated in server, not the same as required.
        json_ret = response.json()
        self.assertEqual(json_ret.get('author', ''), book['author'])
        self.assertEqual(json_ret.get('title', ''), book['title'])
        return json_ret

    def verify_book(self, book):
        print 'Remove book: /shelves/%d/books/%d' % (book['shelf'], book['id'])
        # Get book: api_key
        r = self._send_request(
            '/shelves/%d/books/%d' % (book['shelf'], book['id']),
            api_key=FLAGS.api_key)
        self.assertEqual(r.status_code, 200)
        self.assertEqual(r.json(), book)

    def delete_book(self, book):
        book_name = 'shelves/%d/books/%d' % (book['shelf'], book['id'])
        print 'Remove book: /%s' % book_name
        # Delete book: api_key
        r = self._send_request(
            '/' + book_name, api_key=FLAGS.api_key, method='DELETE')
        self.assertEqual(r.status_code, 204)

    def verify_list_books(self, shelf, books):
        # List book: api_key
        response = self._send_request(
            '/shelves/%d/books' % shelf['id'], api_key=FLAGS.api_key)
        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json().get('books', []), books)

    def run(self):
        shelf1 = {
            'theme': 'Text books'
        }
        shelf1 = self.create_shelf(shelf1)
        shelf2 = {
            'theme': 'Fiction'
        }
        shelf2 = self.create_shelf(shelf2)
        self.verify_shelf(shelf1)
        self.verify_shelf(shelf2)
        book11 = {
            'author': 'Graham Doggett',
            'title': 'Maths for Chemist'
        }
        book11 = self.create_book(shelf1, book11)
        self.verify_list_books(shelf1, [book11])
        book12 = {
            'author': 'George C. Comstock',
            'title': 'A Text-Book of Astronomy'
        }
        book12 = self.create_book(shelf1, book12)
        self.verify_list_books(shelf1, [book11, book12])
        self.verify_book(book11)
        self.verify_book(book12)
        # shelf2 is empty
        self.verify_list_books(shelf2, [])
        self.delete_book(book12)
        self.verify_list_books(shelf1, [book11])
        self.delete_book(book11)
        self.verify_list_books(shelf1, [])
        self.delete_shelf(shelf2)
        self.delete_shelf(shelf1)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(
        description=__doc__,
        formatter_class=argparse.RawTextHelpFormatter)
    parser.add_argument('--verbose', type=bool, help='Turn on/off verbosity.')
    parser.add_argument('--host', help='Deployed application host name.')
    parser.add_argument('--count', help='Number of times to run client logic '
                                        'loop. If no count provided, it will '
                                        'run indefinitely.')
    parser.add_argument('--api_key', help='An API key to use for requests.')
    parser.add_argument('--verify', dest='verify', action='store_true',
                        help='Verify API responses.')
    parser.add_argument('--no-verify', dest='verify', action='store_false',
                        help='Do not verify API responses.')
    parser.set_defaults(verify=False)
    flags = parser.parse_args(namespace=FLAGS)
    bookstore_client = BookstoreClient(flags.verify)

    if flags.count:
      for i in xrange(int(flags.count)):
          bookstore_client.run()
    else:
        while True:
          bookstore_client.run()
