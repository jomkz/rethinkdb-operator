// Copyright 2018 The rethinkdb-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Client to verify connection to a RethinkDB cluster.

r = require('rethinkdb');
fs = require('fs');

const SERVER_HOST = 'rethinkdb-basic-example.rethinkdb.svc.cluster.local';
const SERVER_PORT = 28015;
const SERVER_TIMEOUT = 10;
const SERVER_PASSWORD = fs.readFileSync('/etc/rethinkdb/credentials/admin-password', 'utf8');

r.connect({
    host: SERVER_HOST,
    port: SERVER_PORT,
    timeout: SERVER_TIMEOUT,
    password: SERVER_PASSWORD,
    ssl: {
        'ca': fs.readFileSync('/etc/rethinkdb/tls/ca.crt', 'utf8'),
        'cert': fs.readFileSync('/etc/rethinkdb/tls/client.crt', 'utf8'),
        'key': fs.readFileSync('/etc/rethinkdb/tls/client.key', 'utf8')
    }
}, function (err, conn) {
    if (err) throw err;
    r.db('rethinkdb').table('server_status').pluck('name').run(conn, function (err, res) {
        if (err) throw err;
        res.toArray(function (err, results) {
            if (err) throw err;
            console.log(results);
            process.exit();
        });
    });
});
