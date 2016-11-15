////////////////////////////////////////////////////////////////////////////////
//
// An example implementation of a simple bookstore API.

'use strict';

var express = require('express');
var bodyParser = require('body-parser');


/**
 * @typedef {Object} InitializationOptions
 * @property {Boolean} log Log incoming requests.
 * @property {String} host MySQL backend host name.
 * @property {String} port MySQL backend port.
 * @property {String} user MySQL backend user name.
 * @property {String} password MySQL backend user password.
 * @property {String} database MySQL backend database name.
 */

/**
 * Creates an Express.js application which implements a Bookstore
 * API defined in `swagger.json`.
 *
 * @param {InitializationOptions} options Application initialization options.
 * @return {!express.Application} An initialized Express.js application.
 *
 * If no options are provided, defaults are:
 *     {
 *       log: true,
 *     }
 */
function bookstore(options) {
  options = options || {
    log: true,
  };

  var database = createDatabase(options);

  var app = express();
  if (options.log) {
    app.use(function(req, res, next) {
      console.log(req.method, req.originalUrl);
      next();
    });
  }
  app.use(bodyParser.json());

  // Serve application version for tests to ensure that
  // bookstore was deployed correctly.
  app.get('/version', function(req, res) {
    res.set('Content-Type', 'application/json');
    res.status(200).send({
      version: '${VERSION}'
    });
  });

  // Middleware which returns an error if there is no
  // database connection.
  app.use(function(req, res, next) {
    if (! database) {
      return error(res, 500, "No database connection");
    }
    next();
  });

  /**
   * @typedef {Object} UserInfo
   * @property {String} id An auth provider defined user identity.
   * @property {String} email An authenticated user email address.
   * @property {Object} consumer_id A consumer identifier (currently unused).
   */

  function error(res, status, message) {
    res.status(status).json({
      error: status,
      message: message
    });
  }

  app.get('/shelves', function(req, res) {
    database.listShelves(function(err, shelves) {
      if (err) {
        return error(res, err.error, err.message);
      }

      res.status(200).json({
        shelves: shelves
      });
    });
  });

  app.post('/shelves', function(req, res) {
    var shelfRequest = req.body;
    if (shelfRequest === undefined) {
      return error(res, 400, 'Missing request body.');
    }
    if (shelfRequest.theme === undefined) {
      return error(res, 400, 'Shelf resource is missing required \'theme\'.');
    }

    database.createShelf(shelfRequest.theme, function(err, shelf) {
      res.status(200).json({
        id: shelf.id,
        theme: shelf.theme
      });
    });
  });

  app.get('/shelves/:shelf', function(req, res) {
    database.getShelf(req.params.shelf, function(err, shelf) {
      if (err) {
        return error(res, err.error, err.message);
      }

      res.status(200).json({
        id: shelf.id,
        theme: shelf.theme
      });
    });
  });

  app.delete('/shelves/:shelf', function(req, res) {
    database.deleteShelf(req.params.shelf, function(err) {
      if (err) {
        return error(res, err.error, err.message);
      }

      res.status(204).end();
    });
  });

  app.get('/shelves/:shelf/books', function(req, res) {
    database.listBooks(req.params.shelf, function(err, books) {
      if (err) {
        return error(res, err.error, err.message);
      }

      res.status(200).json({
        books: books
      });
    });
  });

  app.post('/shelves/:shelf/books/', function(req, res) {
    var bookRequest = req.body;
    if (bookRequest === undefined) {
      return error(res, 400, 'Missing request body.');
    }
    if (bookRequest.author === undefined) {
      return error(res, 400, 'Book resource is missing required \'author\'.');
    }
    if (bookRequest.title === undefined) {
      return error(res, 400, 'Book resource is missing required \'title\'.');
    }
    var book = database.createBook(req.params.shelf,
                                   bookRequest.author,
                                   bookRequest.title, function(err, book) {
      if (err) {
        return error(res, err.error, err.message);
      }

      res.status(200).json({
        id: book.id,
        shelf: book.shelf,
        author: book.author,
        title: book.title
      });
    });
  });

  app.get('/shelves/:shelf/books/:book', function(req, res) {
    database.getBook(req.params.shelf, req.params.book, function(err, book) {
      if (err) {
        return error(res, err.error, err.message);
      }

      res.status(200).json({
        id: book.id,
        shelf: book.shelf,
        author: book.author,
        title: book.title
      });
    });
  });

  app.delete('/shelves/:shelf/books/:book', function(req, res) {
    database.deleteBook(req.params.shelf, req.params.book, function(err, book) {
      if (err) {
        return error(res, err.error, err.message);
      }

      res.status(204).end();
    });
  });

  function createInMemoryDatabase() {

    // The bookstore example uses a simple, in-memory database
    // for illustrative purposes only.
    function inMemoryDatabase() {
      this.shelves = {};
      this.id = 0;

      var db = this;

      db.createShelf('Fiction', function(err, fiction) {
        db.createBook(fiction.id, 'Neal Stephenson', 'REAMDE', function(){});
      });
      db.createShelf('Fantasy', function(err, fantasy) {
        db.createBook(fantasy.id, 'George R.R. Martin', 'A Game of Thrones',
                      function(){});
      });
    }

    inMemoryDatabase.prototype.listShelves = listShelves;
    inMemoryDatabase.prototype.createShelf = createShelf;
    inMemoryDatabase.prototype.getShelf = getShelf;
    inMemoryDatabase.prototype.deleteShelf = deleteShelf;
    inMemoryDatabase.prototype.listBooks = listBooks;
    inMemoryDatabase.prototype.createBook = createBook;
    inMemoryDatabase.prototype.getBook = getBook;
    inMemoryDatabase.prototype.deleteBook = deleteBook;

    function listShelves(next) {
      var result = [];
      var shelves = this.shelves;

      for (var id in shelves) {
        var shelf = shelves[id];
        result.push({
          id: shelf.id,
          theme: shelf.theme
        });
      }
      next(undefined, result);
    }

    function createShelf(theme, next) {
      var id = ++this.id;
      var shelf = {
        id: id,
        theme: theme,
        books: {}
      };
      this.shelves[shelf.id] = shelf;
      next(undefined, shelf);
    }

    function getShelf(id, next) {
      var shelf = this.shelves[id];
      if (shelf === undefined) {
        return next({ error: 404, message: 'Shelf ' + id + ' not found.'});
      }
      next(undefined, shelf);
    }

    function deleteShelf(id, next) {
      var shelf = this.shelves[id];
      if (shelf === undefined) {
        return next({ error: 404, message: 'Shelf ' + id + ' not found.'});
      }
      delete this.shelves[id];
      next(undefined);
    }

    function listBooks(shelf, next) {
      var shelf = this.shelves[shelf];
      if (shelf === undefined) {
        return next({ error: 404, message: 'Shelf ' + shelf + ' not found.'});
      }

      var result = [];
      var books = shelf.books;
      for (var id in books) {
        var book = books[id];
        result.push({
          id: book.id,
          shelf: book.shelf,
          author: book.author,
          title: book.title
        });
      }

      next(undefined, result);
    }

    function createBook(shelfName, author, title, next) {
      var shelf = this.shelves[shelfName];
      if (shelf === undefined) {
        return next({
          error: 404,
          message: 'Shelf ' + shelfName + ' not found.'
        });
      }
      var id = ++this.id;
      var book = {
        id: id,
        shelf: shelf.id,
        author: author,
        title: title
      };
      shelf.books[book.id] = book;
      next(undefined, book);
    }

    function getBook(shelfName, bookName, next) {
      var shelf = this.shelves[shelfName];
      if (shelf === undefined) {
        return next({
          error: 404,
          message: 'Shelf ' + shelfName + ' not found.'
        });
      }
      var book = shelf.books[bookName];
      if (book === undefined) {
        return next({ error: 404, message: 'Book ' + bookName + ' not found.'});
      }
      next(undefined, book);
    }

    function deleteBook(shelfName, bookName, next) {
      var shelf = this.shelves[shelfName];
      if (shelf === undefined) {
        return next({
          error: 404,
          message: 'Shelf ' + shelfName + ' not found.'
        });
      }
      var book = shelf.books[bookName];
      if (book === undefined) {
        return next({ error: 404, message: 'Book ' + bookName + ' not found.'});
      }
      delete shelf.books[bookName];
      next(undefined, book);
    }

    return new inMemoryDatabase();
  }

  function createMySQLDatabase(options) {
    // No host was provided, we cannot connect to the database.
    if (!options.host) {
      return null;
    }

    var mysql = require('mysql');

    function MySQLDatabase() {
      var connectionOptions = {
        host    : options.host,
        port    : options.port || 3306,
        user    : options.user,
        password: options.password,
        database: options.database || 'bookstore',
        multipleStatements: true,
      };
      console.log(connectionOptions);

      var database = this;  // For closures.

      function connect() {
        var connection = mysql.createConnection(connectionOptions);

        connection.connect(function(err) {
          if (err) {
            database.connection = undefined;
            console.error('Cannot connect to database ', connectionOptions);
            console.log(err);
            setTimeout(connect, 5000);
          } else {
            console.log('Database connection established.');
            database.connection = connection;
          }
        });

        connection.on('error', function(err) {
          console.log('Database error', err);
          if (err.code === 'PROTOCOL_CONNECTION_LOST') {
            connect();
          } else {
            throw err;
          }
        });
      }

      connect();
    }

    MySQLDatabase.prototype.listShelves = listShelves;
    MySQLDatabase.prototype.createShelf = createShelf;
    MySQLDatabase.prototype.getShelf = getShelf;
    MySQLDatabase.prototype.deleteShelf = deleteShelf;
    MySQLDatabase.prototype.listBooks = listBooks;
    MySQLDatabase.prototype.createBook = createBook;
    MySQLDatabase.prototype.getBook = getBook;
    MySQLDatabase.prototype.deleteBook = deleteBook;

    function listShelves(next) {
      var query = 'CALL list_shelves';
      var resultSet = {
        Shelves: 0,
        OkPacket: 1,
      };

      this.connection.query(query, function(err, results) {
        if (err) {
          return next({error: 500, message: err.message});
        }

        var shelves = [];
        var data = results[resultSet.Shelves];
        for (var i in data) {
          var row = data[i];
          shelves.push({id: parseInt(row.id), theme: row.theme});
        }

        next(undefined, shelves);
      });
    }

    function createShelf(theme, next) {
      var query = 'CALL create_shelf(?, @id); SELECT @id as id;';
      var resultSet = {
        OkPacket: 0,
        ID: 1,
      };

      this.connection.query(query, [theme], function(err, results) {
        if (err) {
          return next({error: 500, message: err.message});
        }

        var idRow = results[resultSet.ID][0];
        next(undefined, {id: parseInt(idRow.id), theme: theme});
      });
    }

    function getShelf(id, next) {
      var query = 'CALL get_shelf(?)'
      var resultSet = {
        Shelf: 0,
        OkPacket: 1,
      };

      this.connection.query(query, [id], function(err, results) {
        if (err) {
          return next({error: 500, message: err.message});
        }

        var shelf = results[resultSet.Shelf][0];
        if (shelf === undefined) {
          return next({
            error: 404,
            message: 'Shelf ' + id + ' not found.'
          });
        }
        next(undefined, {id: parseInt(shelf.id), theme: shelf.theme});
      });
    }

    function deleteShelf(id, next) {
      var query = 'CALL delete_shelf(?, @valid); SELECT @valid as valid;';
      var resultSet = {
        OkPacket: 0,
        Valid: 1,
      };

      this.connection.query(query, [id], function(err, results) {
        if (err) {
          return next({error: 500, message: err.message});
        }

        var validRow = results[resultSet.Valid][0];
        if (! validRow.valid) {
          return next({error: 404, message: 'Shelf ' + id + ' not found.'});
        }

        next(undefined);
      });
    }

    function listBooks(shelf, next) {
      var query = 'CALL list_books(?, @valid); SELECT @valid as valid;';
      var resultSet = {
        Books: 0,
        OkPacket: 1,
        Valid: 2,
      };

      this.connection.query(query, [shelf], function(err, results) {
        if (err) {
          return next({error: 500, message: err.message});
        }

        var validRow = results[resultSet.Valid][0];
        if (! validRow.valid) {
          return next({error: 404, message: 'Shelf ' + shelf + ' not found.'});
        }

        var books = [];
        var data = results[resultSet.Books];
        for (var i in data) {
          var row = data[i];
          books.push({
            id: parseInt(row.id),
            shelf: parseInt(row.shelf),
            author: row.author,
            title: row.title
          });
        }

        next(undefined, books);
      });
    }

    function createBook(shelfName, author, title, next) {
      var query = 'CALL create_book(?, ?, ?, @valid, @id); ' +
                  'SELECT @valid as valid, @id as id;'
      var resultSet = {
        OkPacket: 0,
        ValidAndId: 1,
      };

      this.connection.query(query, [shelfName, author, title],
                            function(err, results) {
        if (err) {
          console.log(err);
          return next({error: 500, message: err.message});
        }

        var validAndIdRow = results[resultSet.ValidAndId][0];

        if (! validAndIdRow.valid) {
          return next({error: 404, message: 'Shelf ' + shelfName + ' not found.'});
        }

        next(undefined, {
            id: parseInt(validAndIdRow.id),
            shelf: parseInt(shelfName),
            author: author,
            title: title
        });
      });
    }

    function getBook(shelfName, bookName, next) {
      var query = 'CALL get_book(?, ?)';
      var resultSet = {
        Books: 0,
        OkPacket: 1,
      };

      this.connection.query(query, [shelfName, bookName],
                            function(err, results) {
        if (err) {
          return next({error: 500, message: err.message});
        }

        var book = results[resultSet.Books][0];

        if (book === undefined) {
          return next({
              error: 404,
              message: 'Book ' + bookName + ' not found.'
          });
        }

        next(undefined, {
            id: parseInt(book.id),
            shelf: parseInt(book.shelf),
            author: book.author,
            title: book.title,
        });
      });
    }

    function deleteBook(shelfName, bookName, next) {
      var query = 'CALL delete_book(?, ?, @valid); SELECT @valid as valid;';
      var resultSet = {
        OkPacket: 0,
        Valid: 1,
      };

      this.connection.query(query, [shelfName, bookName],
                            function(err, results) {
        if (err) {
          return next({error: 500, message: err.message});
        }

        var result = results[resultSet.Valid][0];

        if (! result.valid) {
          return next({error: 404, message: 'Shelf ' + shelfName + ' not found.'});
        }

        next(undefined);
      });
    }

    return new MySQLDatabase(options);
  }

  function createDatabase(options) {
    if (options.mysql) {
      console.log('Creating a MySQL database.');
      return createMySQLDatabase(options.mysql);
    } else {
      console.log('Creating an in-memory database.');
      return createInMemoryDatabase();
    }
  }

  return app;
}

// If this file is imported as a module, export the `bookstore` function.
// Otherwise, if `bookstore.js` is executed as a main program, start
// the server and listen on a port.
if (module.parent) {
  module.exports = bookstore;
} else {
  var port = process.env.PORT || '8080';
  var options = {
    log: true,
  };

  // Use in-memory database only if --memory is present.
  if (process.argv.indexOf('--memory') < 0) {
    // Use MySQL by default.
    options.mysql = {
      host: process.env.MYSQL_HOST || undefined,
      port: process.env.MYSQL_PORT || undefined,
      user: process.env.MYSQL_USER || undefined,
      password: process.env.MYSQL_PASSWORD || undefined,
      database: process.env.MYSQL_DATABASE || undefined,
    }
  }

  var server = bookstore(options).listen(port, '0.0.0.0',
      function() {
        var host = server.address().address;
        var port = server.address().port;

        console.log('Bookstore listening at http://%s:%s', host, port);
      }
  );
}
