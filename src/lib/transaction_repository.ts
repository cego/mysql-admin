import { Instance } from '@/lib/config'
import mysql from 'mysql2/promise'
import { RowDataPacket } from 'mysql2/promise'
import * as processListJson from '@/lib/processlist.json'

export type TransactionInfoDict = {
    [threadId: number]: TransactionInfo
}

export type TransactionInfo = {
    activeTime: number
    info: string[]
}

export type Process = {
    Id: number
    User: string
    Host: string
    db: string | null
    Command: string
    Time: number
    State: string
    Info: string
    Progress: number
}

export type ProcessWithTransaction = Process & {
    transaction: TransactionInfo | null
}

const parseInnoDbStatus = (innoDbStatus: string): TransactionInfoDict => {
    const splitInnoDbStatus = innoDbStatus.split('\n') // Find the line LIST OF TRANSACTIONS FOR EACH SESSION:\n
    const transactionsStartIndex = splitInnoDbStatus.findIndex((line) => line.includes('LIST OF TRANSACTIONS FOR EACH SESSION:'))

    // After the transactionStartIndex, read transactions lines, splitting by lines starting with ---TRANSACTION, until we meet the line --------\n
    const transactions: TransactionInfoDict = {}

    let transaction: TransactionInfo = {
        activeTime: -1,
        info: [],
    }

    for (let i = transactionsStartIndex; i < splitInnoDbStatus.length; i++) {
        const line = splitInnoDbStatus[i]

        if (line === undefined) {
            continue
        }

        if (line.startsWith('--------')) {
            break
        }

        if (line.startsWith('---TRANSACTION')) {
            // Get the active time from the format '..., ACTIVE 1 sec'
            const index = line.indexOf(', ACTIVE')
            const activeTime = parseInt(line.slice(index + 8))

            transaction = {
                activeTime,
                info: [],
            }
        }

        if (line.startsWith('MariaDB thread id')) {
            // Get the thread id from the format `MariaDB thread id 3, ...`
            const threadId = parseInt(line.split(' ')[3])

            transactions[threadId] = transaction
        }

        transaction.info.push(line)
    }

    return transactions
}

export const getTransactionsAndInnoDBStatus = async (instance: Instance): Promise<[ProcessWithTransaction[], string]> => {
    return [processListJson.processList, processListJson.innoDbStatus]
    let conn: mysql.Connection | null = null

    let processList: Process[] = []
    let innoDbStatusString = ''

    try {
        conn = await mysql.createConnection(instance)
        const [processListResult] = await conn.query('SHOW PROCESSLIST;')
        processList = processListResult as Process[]
        const [innoDbStatusResult] = await conn.query<RowDataPacket[]>('SHOW ENGINE INNODB STATUS;')

        innoDbStatusString = innoDbStatusResult[0]['Status'] as string
    } finally {
        conn?.end()
    }

    const innoDbStatus = parseInnoDbStatus(innoDbStatusString)

    const processListWithTransaction: ProcessWithTransaction[] = processList.map((process) => {
        const transaction = innoDbStatus[process.Id] || null
        return {
            ...process,
            transaction,
        }
    })

    // Order by transaction.activeTime desc, then by process.Time desc
    processListWithTransaction.sort((a, b) => {
        if (a.transaction && !b.transaction && a.transaction.activeTime > 10) {
            return -1
        }
        if (!a.transaction && b.transaction && b.transaction.activeTime > 10) {
            return 1
        }
        if (a.transaction && b.transaction) {
            return b.transaction.activeTime - a.transaction.activeTime
        }
        return b.Time - a.Time
    })

    return [processListWithTransaction, innoDbStatusString]
}
