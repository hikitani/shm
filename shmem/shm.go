package shmem

import (
	"fmt"
	"syscall"
	"unsafe"
)

// Mode bits for 'shmget'
const (
	IPC_CREAT  = 01000 // Create key if key does not exist
	IPC_EXCL   = 02000 // Fail if key exists
	IPC_NOWAIT = 04000 // Return error on wait
)

// Control commands for 'shmctl'
const (
	IPC_RMID = 0 // Remove identifier.
	IPC_SET  = 1 // Set 'ipc_perm' options
	IPC_STAT = 2 // Get 'ipc_perm' options
	IPC_INFO = 3 // See ipcs
)

const IPC_PRIVATE = 0

type sharedMemory struct {
	b    []byte
	id   uintptr
	addr uintptr
}

func (shm *sharedMemory) Release() error {
	_, _, errno := syscall.Syscall(syscall.SYS_SHMDT, shm.addr, 0, 0)
	if errno != 0 {
		return fmt.Errorf("sys.shmdt: %w", errno)
	}

	return nil
}

func (shm *sharedMemory) ID() uintptr {
	return shm.id
}

func (shm *sharedMemory) Data() []byte {
	return shm.b
}

// Returns slice that points to shared memory by key.
func Get(key uint, size uint) (*sharedMemory, error) {
	shmid, _, errno := syscall.Syscall(syscall.SYS_SHMGET, uintptr(key), uintptr(size), IPC_CREAT|0660)
	if errno != 0 {
		return nil, fmt.Errorf("sys.shmget: %w", errno)
	}

	shmaddr, _, errno := syscall.Syscall(syscall.SYS_SHMAT, shmid, 0, 0)
	if errno != 0 {
		return nil, fmt.Errorf("sys.shmat: %w", errno)
	}

	return &sharedMemory{
		b:  unsafe.Slice((*byte)(unsafe.Pointer(shmaddr)), size),
		id: shmid,
	}, nil
}

func Remove(shm *sharedMemory) error {
	_, _, errno := syscall.Syscall(syscall.SYS_SHMCTL, shm.id, IPC_RMID, 0)
	if errno != 0 {
		return fmt.Errorf("sys.shmctl: %w", errno)
	}

	return nil
}
