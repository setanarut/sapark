# sapark - SAP Collision Detection Simulation

![screenshotsapark](https://github.com/user-attachments/assets/ee8438df-73dd-497c-9bce-1ad1ca0e8ab3)


This project demonstrates a Sweep and Prune (SAP) collision detection simulation using the [Ark](https://github.com/mlange-42/ark) ECS package and [Ebitengine](https://github.com/hajimehoshi/ebiten).

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/setanarut/sapark.git
    ```
    ```sh
    cd sapark
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

## Run

Press 'A' to add more entities to the simulation.

Run the simulation:
```sh
go run -tags tiny main.go
```
