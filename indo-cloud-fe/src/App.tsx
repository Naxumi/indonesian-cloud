import { useState, useEffect } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

interface Task {
  id: string;
  title: string;
  status: string;
  created_at: string;
}

type StatusFilter = 'all' | 'pending' | 'completed';

export function App() {
  const [tasks, setTasks] = useState<Task[]>([])

  
  const [title, setTitle] = useState('')
  const [status, setStatus] = useState('pending')

  
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editTitle, setEditTitle] = useState('')

  
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')

  const fetchTasks = async (filter: StatusFilter = statusFilter) => {
    try {
      const response = await fetch(`http://localhost:8080/api/tasks?status=${filter}`)
      const data = await response.json()
      setTasks(data || [])
    } catch (error) {
      console.error("Gagal mengambil data:", error)
    }
  }

  useEffect(() => {
    fetchTasks()
    
  }, [statusFilter])

  
  const changeFilter = (filter: StatusFilter) => {
    setStatusFilter(filter)
  }

  
  const addTask = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return

    await fetch('http://localhost:8080/api/tasks', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: title, status: status }) 
    })
    setTitle('')
    fetchTasks()
  }

  
  const toggleStatus = async (id: string, currentStatus: string) => {
    const newStatus = currentStatus === 'pending' ? 'completed' : 'pending'
    await fetch(`http://localhost:8080/api/tasks/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ status: newStatus })
    })
    fetchTasks()
  }

  
  const saveEdit = async (id: string) => {
    await fetch(`http://localhost:8080/api/tasks/${id}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ title: editTitle })
    })
    setEditingId(null)
    fetchTasks()
  }

  
  const startEditing = (task: Task) => {
    setEditingId(task.id)
    setEditTitle(task.title)
  }

  
  const deleteTask = async (id: string) => {
    await fetch(`http://localhost:8080/api/tasks/${id}`, {
      method: 'DELETE'
    })
    fetchTasks()
  }

  return (
    <div className="min-h-screen bg-background p-3 sm:p-6 lg:p-8 flex justify-center items-start">
      <Card className="w-full max-w-2xl mx-auto">
        <CardHeader className="border-b">
          <CardTitle className="text-lg sm:text-xl text-center">
            Mini Task Manager
          </CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-5">

          {/* Form Create */}
          <form
            onSubmit={addTask}
            className="grid gap-2 sm:grid-cols-[1fr_auto_auto] sm:gap-3"
          >
            <Input
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Judul tugas baru..."
              required
              className="w-full"
            />

            {/* Dropdown pilihan status */}
            <select
              value={status}
              onChange={(e) => setStatus(e.target.value)}
              aria-label="Status tugas baru"
              className="h-8 w-full sm:w-36 rounded-lg border border-input bg-background px-3 text-sm shadow-xs transition-colors focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50"
            >
              <option value="pending">Pending</option>
              <option value="completed">Completed</option>
            </select>

            <Button type="submit" className="w-full sm:w-auto">
              Tambah
            </Button>
          </form>

          {/* Filter Status: All / Pending / Completed */}
          <div className="flex gap-1.5 rounded-lg border border-border bg-muted/50 p-1">
            {(['all', 'pending', 'completed'] as StatusFilter[]).map((filter) => (
              <Button
                key={filter}
                variant={statusFilter === filter ? "default" : "ghost"}
                size="sm"
                className="flex-1"
                onClick={() => changeFilter(filter)}
              >
                {filter.charAt(0).toUpperCase() + filter.slice(1)}
              </Button>
            ))}
          </div>

          {/* List Tasks */}
          <div className="flex flex-col gap-2.5">
            {tasks.length === 0 && (
              <div className="flex flex-col items-center justify-center gap-1 rounded-lg border border-dashed border-border py-10 text-center">
                <p className="text-sm font-medium text-foreground/80">
                  Tidak ada tugas
                </p>
                <p className="text-xs text-muted-foreground">
                  {statusFilter === 'all'
                    ? 'Tambahkan tugas baru menggunakan form di atas.'
                    : `Belum ada tugas dengan status "${statusFilter}".`}
                </p>
              </div>
            )}

            {tasks.map(task => (
              <div
                key={task.id}
                className="flex flex-col gap-3 rounded-lg border border-border bg-card p-3.5 transition-colors hover:bg-muted/30"
              >
                
                {/* MODE EDIT */}
                {editingId === task.id ? (
                  <div className="flex flex-col gap-2">
                    <Input
                      value={editTitle}
                      onChange={(e) => setEditTitle(e.target.value)}
                      placeholder="Judul tugas"
                      autoFocus
                    />
                    <div className="flex gap-2 justify-end">
                      <Button size="sm" onClick={() => saveEdit(task.id)}>Simpan</Button>
                      <Button size="sm" variant="ghost" onClick={() => setEditingId(null)}>Batal</Button>
                    </div>
                  </div>
                ) : (
                  /* MODE NORMAL */
                  <>
                    <div className="flex items-start justify-between gap-3">
                      <span
                        className={`text-sm font-medium leading-snug break-words min-w-0 ${
                          task.status === 'completed'
                            ? 'line-through text-muted-foreground'
                            : 'text-foreground'
                        }`}
                      >
                        {task.title}
                      </span>
                      <span
                        className={`shrink-0 inline-flex items-center rounded-full px-2 py-0.5 text-[0.7rem] font-medium ${
                          task.status === 'completed'
                            ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                            : 'bg-amber-500/10 text-amber-600 dark:text-amber-400'
                        }`}
                      >
                        {task.status}
                      </span>
                    </div>

                    <div className="flex gap-2 sm:justify-end">
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex-1 sm:flex-initial"
                        onClick={() => startEditing(task)}
                      >
                        Edit
                      </Button>
                      <Button
                        variant={task.status === 'pending' ? "secondary" : "default"}
                        size="sm"
                        className="flex-1 sm:flex-initial"
                        onClick={() => toggleStatus(task.id, task.status)}
                      >
                        {task.status === 'pending' ? 'Selesai' : 'Batal'}
                      </Button>
                      <Button
                        variant="destructive"
                        size="sm"
                        className="flex-1 sm:flex-initial"
                        onClick={() => deleteTask(task.id)}
                      >
                        Hapus
                      </Button>
                    </div>
                  </>
                )}
              </div>
            ))}
          </div>

        </CardContent>
      </Card>
    </div>
  )
}

export default App