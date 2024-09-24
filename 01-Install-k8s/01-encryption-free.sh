#!/bin/bash
ssh-keygen -t rsa  -C "test key" -P ''
for IP in  { 100 101 102 103 };do
        sshpass -p basic123 ssh-copy-id   -o StrictHostKeyChecking=no   12.0.0.$IP
done
