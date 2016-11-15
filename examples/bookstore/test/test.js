////////////////////////////////////////////////////////////////////////////////
//
// A test for the bookstore example.

var bookstore = require('../bookstore.js');
var assert = require('chai').assert;
var http = require('http');

var PORT = 8080;

function request(method, path, body, next) {
  var headers = {};
  if (body !== null) {
    headers['Content-Type'] = 'application/json';
    headers['Content-Length'] = body.length;
    headers['X-Endpoint-API-UserInfo'] = new Buffer(JSON.stringify({
      id: 'myId',
      email: 'myEmail',
      consumer_id: 'customerId'
    })).toString('base64');
  }
  var r = http.request({
    host: 'localhost',
    port: PORT,
    method: method,
    path: path,
    headers: headers,
  }, function(res) {
    var responseBody = null;
    res.setEncoding('utf8');
    res.on('data', function(chunk) {
      if (responseBody === null) {
        responseBody = chunk;
      } else {
        responseBody += chunk;
      }
    });
    res.on('end', function() {
      next(res, responseBody);
    });
  });
  if (body !== null) {
    r.write(body);
  }
  r.end();
}

function assertContentType(headers, contentType) {
  assert.property(headers, 'content-type');
  assert.isString(headers['content-type']);
  assert(
    headers['content-type'] === contentType ||
    headers['content-type'].startsWith(contentType + ';'));
}

function parseJsonBody(headers, body) {
  assertContentType(headers, 'application/json');
  return JSON.parse(body);
}

describe('bookstore', function() {
  var server;

  before(function(done) {
    // Initialize bookstore in quiet mode.
    var options = {
      log: false
    };
    server = bookstore(options).listen(PORT, '0.0.0.0', function() {
      done();
    });
  });

  after(function(done) {
    if (server) {
      server.close(done);
    } else {
      done();
    }
  });

  beforeEach(function(done) {
    // Delete all shelves.
    // In each turn we list all remaining shelves and delete one of them.
    // The algorithm terminates when an empty list of shelves is returned.
    function loop() {
      request('GET', '/shelves', null, function(res, body) {
        assert.equal(res.statusCode, 200, 'list shelves didn\'t return 200');
        var json = parseJsonBody(res.headers, body);
        var shelves = json.shelves;

        if (shelves && shelves.length > 0) {
            request('DELETE', '/shelves/' + shelves[0].id, null, function(res, body) {
              assert.equal(res.statusCode, 204, 'DELETE valid shelf didn\'t return 204');
              assert.equal(body, null);

              // Proceed deleting the next shelf.
              loop();
            });

        } else {
          // All shelves have been deleted.
          done();
        }
      });
    }
    // Initiate the deletion sequence.
    loop();
  });

  function unsupported(path, tests) {
    for (var method in tests) {
      var body = tests[method];
      it(method + ' is not supported', function(done) {
        request(method, path, body, function(res, _) {
          assert.equal(res.statusCode, 404);
          done();
        });
      });
    }
  }

  function createShelf(theme, done) {
    request('POST', '/shelves', JSON.stringify({
        theme: theme
      }), function(res, body) {
      assert.strictEqual(res.statusCode, 200);
      done(parseJsonBody(res.headers, body));
    });
  }

  function createBook(shelf, author, title, done) {
    request('POST', '/shelves/' + shelf + '/books', JSON.stringify({
        author: author,
        title: title
      }), function(res, body) {
      assert.strictEqual(res.statusCode, 200);
      done(parseJsonBody(res.headers, body));
    });
  }

  describe('/', function() {
    unsupported('/', {
      GET: null,
      PUT: '{}',
      POST: '{}',
      PATCH: '{}',
      DELETE: null,
    });
  });

  describe('/shelves', function() {
    var fictionShelf = null;
    var fantasyShelf = null;

    beforeEach(function(done) {
      createShelf('Fiction', function(fiction) {
        fictionShelf = fiction;
        createShelf('Fantasy', function(fantasy) {
          fantasyShelf = fantasy;
          createBook(fiction.id, 'Neal Stephenson', 'Seveneves',
                     function(seveneves) {
            createBook(fantasy.id, 'J. R. R. Tolkien',
                       'The Lord of the Rings', function(lotr) {
              done();
            });
          });
        });
      });
    });

    it('GET returns list of shelves', function(done) {
      request('GET', '/shelves', null, function(res, body) {
        assert.equal(res.statusCode, 200, 'list shelves didn\'t return 200');

        var json = parseJsonBody(res.headers, body);
        assert.property(json, 'shelves');
        assert.isArray(json.shelves);
        json.shelves.forEach(function(shelf) {
          assert.property(shelf, 'id');
          assert.property(shelf, 'theme');
          assert.isNumber(shelf.id);
          assert.isString(shelf.theme);
        });

        assert.sameDeepMembers(json.shelves, [fictionShelf, fantasyShelf]);

        done();
      });
    });

    it('POST creates a new shelf', function(done) {
      request('POST', '/shelves', JSON.stringify({
        theme: 'Nonfiction'
      }), function(res, body) {
        assert.equal(res.statusCode, 200, 'create shelf didn\'t return 200');

        var shelf = parseJsonBody(res.headers, body);
        assert.property(shelf, 'id');
        assert.isNumber(shelf.id);
        assert.propertyVal(shelf, 'theme', 'Nonfiction');

        done();
      });
    });

    unsupported('/shelves', {
      PUT: '{}',
      PATCH: '{}',
      DELETE: null
    });
  });

  describe('/shelves/{shelf}', function() {
    var testShelf = null;

    beforeEach('create test shelf', function(done) {
      createShelf('Poetry', function(shelf) {
        assert.propertyVal(shelf, 'theme', 'Poetry');
        testShelf = shelf;
        done();
      });
    });

    it('GET of a valid shelf returns shelf', function(done) {
      request('GET', '/shelves/' + testShelf.id, null, function(res, body) {
        assert.equal(res.statusCode, 200, 'GET valid shelf didn\'t return 200');

        var shelf = parseJsonBody(res.headers, body);
        assert.deepEqual(shelf, testShelf);

        done();
      });
    });

    it('GET of an invalid shelf returns 404', function(done) {
      request('GET', '/shelves/999999', null, function(res, body) {
        assert.equal(res.statusCode, 404, 'GET invalid shelf didn\'t return 404');

        var error = parseJsonBody(res.headers, body);
        assert.property(error, 'message');

        done();
      });
    });

    it('DELETE of a valid shelf deletes it', function(done) {
      request('DELETE', '/shelves/' + testShelf.id, null, function(res, body) {
        assert.equal(res.statusCode, 204, 'DELETE valid shelf didn\'t return 204');
        assert.equal(body, null);
        done();
      });
    });

    it('DELETE of an invalid shelf returns 404', function(done) {
      request('DELETE', '/shelves/' + testShelf.id, null, function(res, body) {
        assert.equal(res.statusCode, 204, 'DELETE valid shelf didn\'t return 204');
        assert.equal(body, null);

        // Try to delete the same shelf again.
        request('DELETE', '/shelves/' + testShelf.id, null, function(res, body) {
          assert.equal(res.statusCode, 404, 'DELETE invalid shelf didn\'t return 404');

          var error = parseJsonBody(res.headers, body);
          assert.property(error, 'message');

          done();
        });
      });
    });

    unsupported('/shelves/1', {
      PUT: '{}',
      POST: '{}',
      PATCH: '{}',
    });
  });

  describe('/shelves/{shelf}/books', function() {
    var testShelf = null;
    var testKnuth = null;
    var testStroustrup = null;

    beforeEach('create test shelf and book', function(done) {
      createShelf('Computers', function(shelf) {
        assert.propertyVal(shelf, 'theme', 'Computers');
        testShelf = shelf;

        createBook(shelf.id, 'Donald E. Knuth',
          'The Art of Computer Programming', function(book) {
          assert.propertyVal(book, 'author', 'Donald E. Knuth');
          assert.propertyVal(book, 'title', 'The Art of Computer Programming');
          testKnuth = book;

          createBook(shelf.id, 'Bjarne Stroustrup',
            'The C++ Programming Language', function(book) {
            assert.propertyVal(book, 'author', 'Bjarne Stroustrup');
            assert.propertyVal(book, 'title', 'The C++ Programming Language');
            testStroustrup = book;

            done();
          });
        });
      });
    });

    it('GET lists books on a valid shelf', function(done) {
      request('GET', '/shelves/' + testShelf.id + '/books', null, function(res, body) {
        assert.strictEqual(res.statusCode, 200, 'List books didn\'t return 200');

        var response = parseJsonBody(res.headers, body);
        assert.property(response, 'books');
        assert.isArray(response.books);
        assert.sameDeepMembers(response.books, [testKnuth, testStroustrup]);

        done();
      });
    });

    it('GET returns 404 for an invalid shelf', function(done) {
      request('GET', '/shelves/999999/books', null, function(res, body) {
        assert.strictEqual(res.statusCode, 404);

        var error = parseJsonBody(res.headers, body);
        assert.property(error, 'message');

        done();
      });
    });

    it('POST creates a new book in a valid shelf', function(done) {
      var practice = {
        author: 'Brian W. Kernighan, Rob Pike',
        title: 'The Practice of Programming'
      };
      request('POST', '/shelves/' + testShelf.id + '/books', JSON.stringify(practice),
        function(res, body) {
          assert.strictEqual(res.statusCode, 200);

          var book = parseJsonBody(res.headers, body);

          assert.propertyVal(book, 'author', 'Brian W. Kernighan, Rob Pike');
          assert.propertyVal(book, 'title', 'The Practice of Programming');
          assert.property(book, 'id');
          assert.isNumber(book.id);

          done();
        });
    });

    it('POST returns 404 for an invalid shelf', function(done) {
      var compilers = {
        author: 'Aho, Sethi, Ullman',
        title: 'Compilers'
      };
      request('POST', '/shelves/999999/books', JSON.stringify(compilers),
        function(res, body) {
          assert.strictEqual(res.statusCode, 404);

          var error = parseJsonBody(res.headers, body);
          assert.property(error, 'message');

          done();
        });
    });

    unsupported('/shelves/1/books', {
      PUT: '{}',
      PATCH: '{}',
      DELETE: null
    });
  });

  describe('/shelves/{shelf}/books/{book}', function() {
    var testYoga = null;
    var testSutras = null;
    var testBreathing = null;

    beforeEach('create test shelf and books', function(done) {
      createShelf('Yoga', function(shelf) {
        assert.propertyVal(shelf, 'theme', 'Yoga');
        testYoga = shelf;

        createBook(shelf.id, 'Patanjali', 'Yoga Sutras of Patanjali', function(book) {
          assert.propertyVal(book, 'author', 'Patanjali');
          assert.propertyVal(book, 'title', 'Yoga Sutras of Patanjali');
          testSutras = book;

          createBook(shelf.id, 'Donna Farhi', 'The breathing book', function(book) {
            assert.propertyVal(book, 'author', 'Donna Farhi');
            assert.propertyVal(book, 'title', 'The breathing book');
            testBreathing = book;

            done();
          });
        });
      });
    });

    it('GET of a valid book returns a book', function(done) {
      request('GET', '/shelves/' + testBreathing.shelf + '/books/' + testBreathing.id,
              null, function(res, body) {
        assert.strictEqual(res.statusCode, 200);

        var book = parseJsonBody(res.headers, body);
        assert.deepEqual(book, testBreathing);

        done();
      });
    });

    it('GET of a book on an invalid shelf returns 404', function(done) {
      request('GET', '/shelves/999999/books/5', null, function(res, body) {
        assert.strictEqual(res.statusCode, 404);

        var error = parseJsonBody(res.headers, body);
        assert.property(error, 'message');

        done();
      });
    });

    it('GET of an invalid book on valid shelf returns 404', function(done) {
      request('GET', '/shelves/' + testYoga.id + '/books/999999', null, function(res, body) {
        assert.strictEqual(res.statusCode, 404);

        var error = parseJsonBody(res.headers, body);
        assert.property(error, 'message');

        done();
      });
    });

    it('DELETE of a valid book deletes the book', function(done) {
      request('DELETE', '/shelves/' + testSutras.shelf + '/books/' + testSutras.id,
              null, function(res, body) {
        assert.strictEqual(res.statusCode, 204);
        assert.equal(body, null);
        done();
      });
    });

    it('DELETE of a book on an invalid shelf returns 404', function(done) {
      request('DELETE', '/shelves/999999/books/5', null, function(res, body) {
        assert.strictEqual(res.statusCode, 404);

        var error = parseJsonBody(res.headers, body);
        assert.property(error, 'message');

        done();
      });
    });

    it('DELETE of an invalid book on a valid shelf returns 404', function(done) {
      // Delete the book as above.
      request('DELETE', '/shelves/' + testSutras.shelf + '/books/' + testSutras.id,
              null, function(res, body) {
        assert.strictEqual(res.statusCode, 204);
        assert.equal(body, null);
        // Delete the same book again.
        request('DELETE', '/shelves/' + testSutras.id, null, function(res, body) {
          assert.strictEqual(res.statusCode, 404);

          var error = parseJsonBody(res.headers, body);
          assert.property(error, 'message');

          done();
        });
      });
    });

    unsupported('/shelves/1/books/2', {
      PUT: '{}',
      POST: '{}',
      PATCH: '{}',
    });
  });

  describe('/version', function() {
    it('GET returns version', function(done) {
      request('GET', '/version', null, function(res, body) {
        assert.equal(res.statusCode, 200, '/version didn\'t return 200');

        var json = parseJsonBody(res.headers, body);
        assert.property(json, 'version');
        assert.isString(json.version);

        done();
      });
    });
  });
});

describe('bookstore MySQL', function() {
  var server;

  before(function(done) {
    // Initialize bookstore in quiet mode.
    var options = {
      log: false,
      mysql: {},
    };
    server = bookstore(options).listen(PORT, '0.0.0.0', function() {
      done();
    });
  });

  after(function(done) {
    if (server) {
      server.close(done);
    } else {
      done();
    }
  });

  describe('SQL without connection information', function() {
    it('/version returns success', function(done) {
      request('GET', '/version', null, function(res, body) {
        assert.equal(res.statusCode, 200, '/version didn\'t return 200');
        done();
      });
    });

    tests = [
      [ 'GET', '/shelves' ],
      [ 'POST', '/shelves', JSON.stringify({ theme: 'Travel' }) ],
      [ 'GET', '/shelves/1' ],
      [ 'DELETE', '/shelves/1' ],
      [ 'GET', '/shelves/1/books' ],
      [ 'POST', '/shelves/1/books/', JSON.stringify({
        author: 'Rick Steves', title: 'Travel as a Political Act'
      })],
      [ 'GET', '/shelves/1/books/1' ],
      [ 'DELETE', '/shelves/1/books/1' ],
    ];

    for (i in tests) {
      var t = tests[i];
      (function(verb, url, body) {
        it(verb + ' ' + url + ' returns 500', function(done) {
          request(verb, url, body, function(res, body) {
            assert.equal(res.statusCode, 500, verb + ' ' + url + ' didn\'t return 500');
            done();
          });
        });
      })(t[0], t[1], t.length > 2 ? t[2] : null);
    }
  });
});
