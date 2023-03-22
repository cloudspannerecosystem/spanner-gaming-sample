# Workload

You can run these workloads either locally or on GKE.

Either way before the workloads can be used, first deploy the services according to the [README](../README.md).

> **NOTE:** The GKE instructions also include deploying the workloads on GKE.

The rest of this document explains how to run the workloads locally.

### Workload dependencies

There are several dependencies required to get the generators to work:

- Python 3.7+
- Locust

Assuming python3.X is installed, install dependencies via [pip](https://pypi.org/project/pip/):

```
# if pip3 is symlinked to pip
pip install -r requirements.txt

# if pip3 is not symlinked to pip
pip3 install -r requirements.txt
```

> **NOTE:** To avoid modifying existing pip libraries on your machine, consider a solution like [virtualenv](https://pypi.org/project/virtualenv/).


## Using the workload generators
The provided workload generators do the following:

- _authentication\_server.py_: mimics player signup, player logins, player retrieval by UUID and player logouts

Run on the CLI:
```
locust -H http://127.0.0.1:8080 -f ./workloads/profile/authentication_server.py --headless -u=2 -r=2 -t=10s
```

Run webUI on port 8090:
```
locust --web-port 8090 -f ./workloads/profile/authentication_server.py
# Connect browser to http://localhost:8090
```

- _match\_server.py_: mimics game servers matching players together, and closing games out.

Run on the CLI:
```
locust -H http://127.0.0.1:8081 -f ./workloads/matchmaking/match_server.py --headless -u=1 -r=1 -t=10s
```

Run on port 8091:
```
locust --web-port 8091 -f ./workloads/matchmaking/match_server.py
# Connect browser to http://localhost:8091
```

- _item\_generator.py_: generates some random items for our game to use.

Run on the CLI for 1 minute:
```
locust -H http://127.0.0.1:8082 -f ./workloads/item_generator.py --headless -u=1 -r=1 -t=60s
```

- _game\_server.py_: mimics adding loot and money to players during the course of a game

Run on the CLI:
```
locust -H http://127.0.0.1:8082 -f ./workloads/game/game_server.py --headless -u=1 -r=1 -t=10s
```

Run on port 8092:
```
locust --web-port 8092 -f ./workloads/game/game_server.py
# Connect browser to http://localhost:8092
```

- _trading\_server.py_: mimics posting and closing orders on trading post

Run on the CLI:
```
locust -H http://127.0.0.1:8083 -f ./workloads/tradepost/trading_server.py --headless -u=1 -r=1 -t=10s
```

Run on port 8093:
```
locust --web-port 8093 -f ./workloads/tradepost/trading_server.py
# Connect browser to http://localhost:8093
```
