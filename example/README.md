# Examples

```bash
ecrgate --repo joel-test
```
- Use local dir as Dockerfile path
- Push image with tag latest
- Use default threshold levels

&nbsp;

```bash
ecrgate --repo joel-test --dockerfile ubuntu/ --tag ubuntu --clean
```
- Use ubuntu/ as Dockerfile path
- Push image with tag ubuntu
- Use default threshold levels
- Purge image from ecr repo if scan fails threshold

&nbsp;

```bash
ecrgate --repo joel-test --dockerfile ubuntu/ --tag ubuntu --clean \
    --info 10 --low 5 --medium 3 --high 2 --critical 1
```
- Use ubuntu/ as Dockerfile path
- Push image with tag ubuntu
- Use specified threshold levels
- Purge image from ecr repo if scan fails threshold