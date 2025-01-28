# saparche - SAP Collision Detection Simulation

This project demonstrates a Sweep and Prune (SAP) collision detection simulation using the Arche ECS package and Ebitengine.

## Features

- **Entity-Component-System (ECS)**: Utilizes the Arche ECS package for efficient entity management.
- **Collision Detection**: Implements the SAP algorithm for detecting collisions between entities.
- **Dynamic Entities**: Press 'A' to add more entities to the simulation.
- **Boundary Handling**: Entities bounce off the screen boundaries.

## Requirements

- Go 1.23.5 or later

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/setanarut/saparche.git
    ```
    ```sh
    cd saparche
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

## Usage

Run the simulation:
```sh
go run main.go
```
